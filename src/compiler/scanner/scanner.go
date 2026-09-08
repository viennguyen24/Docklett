package scanner

/*
	Scanner reads the input Dockerfile code and produces a list of tokens for the parser to consume.
	A Lexme is a literal string from the file, literally. A Token is a wrapper around the Lexme raw string and add internal information useful parsing.

	A scanner job is to use a fixed set of text-matching rules we define, in correct order, to process all characters in an input text to a list of tokens. The token produced must depend on correct mode / context (under what set of rules the token is applying to the text in correct order of priority), includes all characters possibles that belong to that token logically, do not include invalid stray characters or characters belonging to another token. In a sense, its job is text processing based on position and rules - it is agnostic about what grammar applies to those token list.

	Two syntaxes share one file, so the scanner must know which rule set owns the next character.

	1. Vanilla Docker instruction
	   Shape: <instruction verb> <everything else on that logical instruction>
	   Verb is a known Docker keyword (FROM, RUN, COPY, …), matched case-insensitively.
	   "Everything else" is opaque: shell text, builder flags (--mount=…), paths, JSON, quotes,
	   @looking package names, ${placeholders}, inline #, and physical newlines that belong to
	   the same instruction because of \ (or ` after # escape=`) continuation or a heredoc body.
	   We emit only two tokens for that instruction: DOCKER_KEYWORD + DOCKER_ARGS.
	   We do not sub-lex the args into Docklett operators or directives.

	2. Custom Docklett instruction
	   Shape: "@" + <Docklett keyword> then expression pieces on that line
	   Keywords come from DocklettTokenKeywords (@SET, @IF, @FOR, …; bare TRUE/FALSE/IN/range while
	   still on that Docklett line). After the keyword we emit normal expression tokens: numbers,
	   strings, identifiers, operators, (), [], commas. Unknown @Name is a ScanError.

	Mode / boundary intuition:
	- Default at the start of a logical line is Docker-oriented: a known verb starts a vanilla
	  instruction and owns args until that instruction is finished.
	- "@" at line start flips docklett mode for the rest of that line so the following words use
	  Docklett keyword/expression rules.
	- NLINE clears docklett mode and returns to "instruction may start here" for the next line.
	- We only accept a new instruction (Docker verb or @directive) when atLineStart is true.
	  Inside an unfinished Docker args region, a physical newline is still args, not a new line start.
	- Docker names used mid-Docklett-expression (e.g. @SET RUN = 1) stay DOCKER_KEYWORD for
	  reservation, but we do not switch into Docker-args mode or swallow "= 1".

	A newline inside a Docker continuation or heredoc must not start a Docklett directive.
	We read that whole argument region before returning to instruction-start mode, so @SET inside it stays data.
	For Docklett source, we keep individual lexemes and report characters or literals we cannot scan.
	We do not evaluate expressions, interpret Docker arguments, or choose tokens based on translator behavior.
*/

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	compileError "docklett/compiler/error"
	"docklett/compiler/token"
)

// We keep source offsets, walk cursor, and line mode together as we walk the input.
// These fields are scanner machine state, not token fields: a Token only stores the finished
// Type/Lexeme/Literal plus Position (Line/Col/File) copied from lexemeLine/lexemeColumn at emit time.
// Each ScanSource call resets this state; ReadSource only loads the input and its file metadata.
type Scanner struct {
	SourcePath string        // filepath of source code
	SourceName string        // filename of source code
	Source     string        // actual source code
	start      int           // first character of current lexeme
	current    int           // current char in source code
	line       int           // current line in source code
	Tokens     []token.Token // list of tokens generated

	// start and current uses byte offsets to slice the source; column counts runes.
	column         int             // one-based rune column of the cursor while walking
	lexemeLine     int             // one-based line where the current token started
	lexemeColumn   int             // one-based rune column where the current token started
	atLineStart    bool            // no instruction token has started on this logical line
	docklett       bool            // flag for whether we are using Docklett extensions
	escape         rune            // Docker continuation character: backslash or backtick
	directivesOpen bool            // Docker parser directives still apply in the initial preamble
	seenDirectives map[string]bool // parser directive names already used in that preamble
	// holds queued DOCKER_ARGS token between scan cycles
	// this described the original pendingToken field; args are now emitted in the same cycle.
}

// Loads a file into the scanner and fills source metadata.
func (s *Scanner) ReadSource(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	absPath, absErr := filepath.Abs(filename)
	if absErr != nil {
		absPath = filename
	}

	s.SourcePath = filepath.Dir(absPath)
	s.SourceName = filepath.Base(absPath)
	s.Source = string(data)
	return nil
}

// in each iteration we scan 1 token
// We record the token's start before scanning, then emit it before reading any Docker args.
// A reserved Docker name inside Docklett mode stays one keyword token; it cannot consume the rest of the line.
// Any scan error clears the result, and only a completed scan appends EOF.
func (s *Scanner) ScanSource() (err error) {
	s.start, s.current = 0, 0
	s.line, s.column = 1, 1
	s.lexemeLine, s.lexemeColumn = 1, 1
	s.atLineStart, s.docklett = true, false
	s.escape = '\\'
	s.directivesOpen = true
	s.seenDirectives = make(map[string]bool)
	s.Tokens = nil
	defer func() {
		if err != nil {
			s.Tokens = nil
		}
	}()
	// A leading UTF-8 BOM is source metadata, not an instruction character.
	if strings.HasPrefix(s.Source, "\ufeff") {
		s.current = len("\ufeff")
	}

	// drain any pending token before scanning new input
	// args are now emitted with their keyword, so there is no pending token to drain here.
	for !s.isAtEnd() {
		s.start = s.current // begin new lexeme
		s.lexemeLine, s.lexemeColumn = s.line, s.column
		instructionStart := s.atLineStart
		tokenType, literal, err := s.scanToken()
		if err != nil {
			return err
		}
		if tokenType == token.ILLEGAL {
			continue
		}
		s.addToken(tokenType, literal)
		if tokenType == token.NLINE {
			continue
		}
		s.atLineStart = false
		s.directivesOpen = false
		// Only an instruction-start keyword owns Docker args; all other tokens are complete here.
		if tokenType != token.DOCKER_KEYWORD || !instructionStart || s.docklett {
			continue
		}
		keyword := s.Source[s.start:s.current]
		args, err := s.scanDockerArgs(keyword)
		if err != nil {
			return err
		}
		// Vanilla Docker arguments come lasts in the original Dockerfile instructions
		s.Tokens = append(s.Tokens, args)
	}

	// drain final pending token after source is fully consumed
	// the last args token is already emitted above, including when the source ends without a newline.
	s.Tokens = append(s.Tokens, token.Token{
		Type: token.EOF,
		Position: token.Position{
			Line: s.line,
			File: s.SourceName,
			Col:  s.column,
		},
	})
	return nil
}

// The byte cursor reaches EOF after every source span has been consumed.
func (s *Scanner) isAtEnd() bool {
	return s.current >= len(s.Source)
}

// Keep the source lexeme and its starting position, even if scanning crossed a physical newline.
func (s *Scanner) addToken(tokenType token.TokenType, literal any) {
	lexeme := s.Source[s.start:s.current]
	s.Tokens = append(s.Tokens, token.Token{
		Type:   tokenType,
		Lexeme: lexeme,
		Position: token.Position{
			Line: s.lexemeLine,
			File: s.SourceName,
			Col:  s.lexemeColumn,
		},
		Literal: literal,
	})
}

// Consume one rune: byte offsets slice the source, while rune columns locate tokens for the user.
func (s *Scanner) advanceChar() rune {
	if s.isAtEnd() {
		return 0
	}
	r, width := utf8.DecodeRuneInString(s.Source[s.current:])
	s.current += width
	if r == '\n' {
		s.line++
		s.column = 1
	} else {
		s.column++
	}
	return r
}

// Read the next rune without consuming it; callers use isAtEnd to distinguish EOF from a zero rune.
func (s *Scanner) peekChar() rune {
	if s.isAtEnd() {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(s.Source[s.current:])
	return r
}

// Report the start of the lexeme or Docker instruction whose scanning failed.
func (s *Scanner) scanError(message string) error {
	return compileError.NewScanError(s.lexemeLine, s.lexemeColumn, s.SourceName, message)
}

// Consume the next rune only when it completes the token we are looking for.
func (s *Scanner) nextMatch(expected rune) bool {
	if s.isAtEnd() {
		return false
	}
	if s.peekChar() != expected {
		return false
	}
	s.advanceChar()
	return true
}

// main logic to process tokens
// generally the design is we lex vanilla Dockerfile instructions separately from new Docklett syntax we define.
// Vanilla Docker instructions is <instruction> <args for instructions>, we process the entire instruction + args.
// custom Docklett instruction is "@" + <Docklett instruction> <"args" like expressions, numbers, strings, identifiers etc>
// so generally, string -> Dockerfile vanilla ? entire vanilla instruction is preserved into instruction and args (only 2 strings, nothing more) -> else try to apply docklett instructions -> each specific arg has their own token
func (s *Scanner) scanToken() (tokenType token.TokenType, literal any, err error) {
	lexeme := s.advanceChar()
	switch lexeme {
	// skip whitespace/newlines without emitting tokens
	// whitespace is skipped here; the newline case below emits NLINE.
	case ' ', '\t', '\r':
		return token.ILLEGAL, nil, nil
	case '\n':
		return s.scanNewline()

	case '=':
		return s.scanOperator('=', token.ASSIGN, token.EQUAL)
	case '+':
		return s.scanOperator('=', token.ADD, token.ADD_ASSIGN)
	case '-':
		return s.scanOperator('=', token.SUBTRACT, token.SUB_ASSIGN)
	case '*':
		return s.scanOperator('=', token.MULTI, token.MULTI_ASSIGN)
	case '/':
		return s.scanOperator('=', token.DIVIDE, token.DIV_ASSIGN)
	case '<':
		return s.scanOperator('=', token.LESS, token.LTE)
	case '>':
		return s.scanOperator('=', token.GREATER, token.GTE)
	case '!':
		return s.scanOperator('=', token.NEGATE, token.UNEQUAL)
	case '&':
		return s.scanOperator('&', token.ILLEGAL, token.AND)
	case '|':
		return s.scanOperator('|', token.ILLEGAL, token.OR)
	case '(':
		return token.LPAREN, nil, nil
	case ')':
		return token.RPAREN, nil, nil
	case '{':
		return token.LBRACE, nil, nil
	case '}':
		return token.RBRACE, nil, nil
	case '[':
		return token.LBRACKET, nil, nil
	case ']':
		return token.RBRACKET, nil, nil
	case ':':
		return token.COLON, nil, nil
	case ',':
		return token.COMMA, nil, nil
	case '#':
		// only ignore full line comments for now
		// inline comments goes into the Docker instruction itself
		return s.scanComment()
	case '"':
		return s.scanStringToken()
	case '@':
		return s.scanDirectiveStart()
	}
	if unicode.IsDigit(lexeme) {
		return s.scanNumberToken()
	}
	if unicode.IsLetter(lexeme) {
		return s.scanKeywordsAndIdentifierTokens()
	}
	return token.ILLEGAL, nil, s.scanError(fmt.Sprintf("unexpected char: %q", lexeme))
}

// We reach this newline only after any Docker payload has been consumed.
// A blank line ends the parser-directive preamble; every logical newline resets Docklett mode.
func (s *Scanner) scanNewline() (token.TokenType, any, error) {
	if s.atLineStart {
		s.directivesOpen = false
	}
	s.atLineStart = true
	s.docklett = false
	return token.NLINE, nil, nil
}

// We try the two-character operator first, then use its single-character form.
// ILLEGAL means the single character has no token, as with lone & and |.
// Keeping this decision here preserves the longest-match rule for every operator above.
func (s *Scanner) scanOperator(next rune, single, paired token.TokenType) (token.TokenType, any, error) {
	if s.nextMatch(next) {
		return paired, nil, nil
	}
	if single == token.ILLEGAL {
		return token.ILLEGAL, nil, s.scanError(fmt.Sprintf("unexpected char: %c", next))
	}
	return single, nil, nil
}

// @ enters Docklett mode only at an instruction boundary.
// Another @ in Docklett mode keeps its lexeme for grammar validation; Docker args never reach this helper.
func (s *Scanner) scanDirectiveStart() (token.TokenType, any, error) {
	if !s.atLineStart && !s.docklett {
		return token.ILLEGAL, nil, s.scanError("unexpected @ outside a Docklett directive")
	}
	s.docklett = true
	return s.scanDocklettToken()
}

// when encounter a #, ignore the entire line because it's a comment
func (s *Scanner) scanComment() (tokenType token.TokenType, literal string, error error) {
	for !s.isAtEnd() {
		nextChar := s.peekChar()
		if nextChar == '\n' {
			break
		}
		s.advanceChar()
	}
	// Only a full-line comment can carry a Docker parser directive.
	if !s.atLineStart {
		return token.ILLEGAL, "", nil
	}
	s.atLineStart = false // a comment line does not count as a blank preamble line
	if err := s.scanParserDirective(s.Source[s.start+1 : s.current]); err != nil {
		return token.ILLEGAL, "", err
	}
	return token.ILLEGAL, "", nil
}

// when encounter a ", read every single char next until we meet the ending "
func (s *Scanner) scanStringToken() (tokenType token.TokenType, literal string, error error) {
	valueStart := s.current
	for !s.isAtEnd() {
		nextChar := s.peekChar()
		if nextChar == '"' {
			break
		}
		// read through new lines
		s.advanceChar()
	}
	if s.isAtEnd() {
		return token.ILLEGAL, "", s.scanError("unterminated string literal")
	}
	strLiteral := s.Source[valueStart:s.current]
	s.advanceChar() // consume closing "
	return token.STRING, strLiteral, nil

}

// We consume the numeric span first, then convert it once; failed conversion is a ScanError.
func (s *Scanner) scanNumberToken() (tokenType token.TokenType, literal any, err error) {
	s.scanDigits()
	// floating number case
	isFloat := s.peekChar() == '.'
	if isFloat {
		s.advanceChar() // consume the dot
		s.scanDigits()
	}

	// Integer and decimal forms keep their existing literal types and share the same error check.
	if isFloat {
		literal, err = strconv.ParseFloat(s.Source[s.start:s.current], 64)
	} else {
		literal, err = strconv.Atoi(s.Source[s.start:s.current])
	}
	if err != nil {
		return token.ILLEGAL, nil, s.scanError("invalid or out-of-range number")
	}
	return token.NUMBER, literal, nil
}

// We use the same digit rule before and after the decimal point; the caller handles conversion errors.
func (s *Scanner) scanDigits() {
	for !s.isAtEnd() {
		nextChar := s.peekChar()
		if !unicode.IsDigit(nextChar) {
			break
		}
		s.advanceChar()
	}
}

// Reads chars after a Docker keyword until newline, handling backslash continuations.
// Returns only the argument portion (whitespace-trimmed), not the keyword itself.
// this now includes heredoc bodies too; their whitespace stays in the args token.
// the normalized header is only used to find boundaries; the token keeps the source text.
// We find the header's end, collect its heredoc delimiters, then consume each body in source order.
// Only after every body closes can we emit the original argument span as one token.
func (s *Scanner) scanDockerArgs(keyword string) (token.Token, error) {
	// skip whitespace between keyword and args
	for !s.isAtEnd() {
		nextChar := s.peekChar()
		if nextChar != ' ' && nextChar != '\t' && nextChar != '\r' {
			break
		}
		s.advanceChar()
	}

	argsStart := s.current
	position := token.Position{Line: s.line, Col: s.column, File: s.SourceName}
	header := s.scanDockerHeader()
	heredocs, err := s.dockerHeredocs(keyword, header)
	if err != nil {
		return token.Token{}, s.scanError(err.Error())
	}
	for _, doc := range heredocs {
		if err := s.scanDockerHeredoc(doc); err != nil {
			return token.Token{}, err
		}
	}
	args := s.Source[argsStart:s.current]
	if len(heredocs) == 0 {
		args = strings.TrimSpace(args)
	}
	return token.Token{Type: token.DOCKER_ARGS, Lexeme: args, Position: position}, nil
}

// Read one physical line, keeping its final newline for the caller to consume or emit.
// We strip CR only from this inspection copy; Source still holds the original bytes.
func (s *Scanner) scanPhysicalLine() string {
	lineByteStart := s.current
	for !s.isAtEnd() && s.peekChar() != '\n' {
		s.advanceChar()
	}
	return strings.TrimSuffix(s.Source[lineByteStart:s.current], "\r")
}

// Join continued lines for heredoc detection, without rewriting the argument token.
// Blank and comment lines inside a continuation do not end the instruction.
func (s *Scanner) scanDockerHeader() string {
	var header strings.Builder
	continued := false
	for {
		line := s.scanPhysicalLine()
		trimmed := strings.TrimSpace(line)
		skipLine := continued && (trimmed == "" || strings.HasPrefix(trimmed, "#"))
		// At EOF there is no newline to skip; otherwise keep reading the same continuation.
		if skipLine && s.isAtEnd() {
			break
		}
		if skipLine {
			s.advanceChar()
			continue
		}
		// backslash before newline means the instruction continues on the next line
		// use the configured escape character here, and account for an escaped escape character.
		part, more := s.trimContinuation(line)
		header.WriteString(part)
		if !more || s.isAtEnd() {
			// leave final newline for main scanner to emit NLINE token
			// a heredoc below may consume this newline first to read its body.
			break
		}
		continued = true
		s.advanceChar() // consume newline for continuation
	}
	return header.String()
}

// Read one attached body through its exact delimiter, preserving @ text and whitespace in Source.
// EOF before the delimiter fails the instruction; the closing newline remains for NLINE emission.
func (s *Scanner) scanDockerHeredoc(doc dockerHeredoc) error {
	if s.isAtEnd() {
		return s.scanError(fmt.Sprintf("unterminated heredoc %q", doc.name))
	}
	s.advanceChar() // header newline, or the previous heredoc's closing newline
	for !s.isAtEnd() {
		line := s.scanPhysicalLine()
		if doc.chomp {
			line = strings.TrimLeft(line, "\t")
		}
		if line == doc.name {
			// leave the final newline for NLINE emission
			// returning here replaces the old loop break; the cursor stays at the same newline.
			return nil
		}
		// advanceChar leaves EOF alone, so a missing delimiter falls through to the error below.
		s.advanceChar()
	}
	return s.scanError(fmt.Sprintf("unterminated heredoc %q", doc.name))
}

// Docker parser directives only apply in the initial, uninterrupted comment
// preamble. Keep their source in Source; only escape affects lexical boundaries.
func (s *Scanner) scanParserDirective(comment string) error {
	if !s.directivesOpen {
		return nil
	}
	name, value, found := strings.Cut(strings.TrimSpace(comment), "=")
	name = strings.ToLower(strings.TrimSpace(name))
	value = strings.TrimSpace(value)
	if !found || value == "" || (name != "syntax" && name != "escape" && name != "check") {
		s.directivesOpen = false
		return nil
	}
	if s.seenDirectives[name] {
		return s.scanError(fmt.Sprintf("duplicate Docker parser directive %q", name))
	}
	s.seenDirectives[name] = true
	// Only escape changes scanning state; syntax and check are recorded above for duplicate detection.
	if name != "escape" {
		return nil
	}
	if value != "\\" && value != "`" {
		return s.scanError("Docker escape character must be a backslash or backtick")
	}
	s.escape = rune(value[0])
	return nil
}

// Match Docker's continuation boundary without changing the stored arguments.
// An escape preceded by another escape does not continue the instruction.
func (s *Scanner) trimContinuation(line string) (string, bool) {
	end := len(strings.TrimRight(line, " \t"))
	if end == 0 || line[end-1] != byte(s.escape) {
		return line, false
	}
	if end > 1 && line[end-2] == byte(s.escape) {
		return line, false
	}
	return line[:end-1], true
}

// A heredoc needs a closing word and, for <<-, permission to ignore leading tabs when matching it.
type dockerHeredoc struct {
	name  string // quote-removed closing delimiter
	chomp bool   // <<- allows leading tabs on the closing line
}

// Inspect only the lexical forms needed to locate attached payloads. Flags,
// JSON, and shell words remain in the original DOCKER_ARGS token.
// Boundary reference: BuildKit frontend/dockerfile/parser/parser.go.
func (s *Scanner) dockerHeredocs(keyword, args string) ([]dockerHeredoc, error) {
	args = dockerArgsAfterFlags(args, s.escape)
	// ONBUILD can wrap a Docker instruction; inspect that header without changing the raw args.
	if strings.EqualFold(keyword, "ONBUILD") {
		keyword, args = dockerOnbuildInstruction(args, s.escape)
	}
	switch strings.ToUpper(keyword) {
	case "RUN", "COPY", "ADD":
	default:
		return nil, nil
	}
	if !strings.Contains(args, "<<") || (strings.HasPrefix(args, "[") && json.Valid([]byte(args))) {
		return nil, nil
	}
	var docs []dockerHeredoc
	// BuildKit uses backslash for heredoc word quoting independently of # escape=.
	for _, word := range dockerWords(args, '\\', true) {
		raw := word.raw
		i := 0
		for i < len(raw) && raw[i] >= '0' && raw[i] <= '9' {
			i++ // optional file descriptor, e.g. 3<<EOF
		}
		if !strings.HasPrefix(raw[i:], "<<") {
			continue
		}
		rest := raw[i+2:]
		chomp := strings.HasPrefix(rest, "-")
		if chomp {
			rest = rest[1:]
		}
		// This also excludes here-strings (<<<), which have no attached body.
		if strings.Contains(rest, "<") {
			continue
		}
		delimiters := dockerWords(strings.TrimSpace(rest), '\\', false)
		if len(delimiters) != 1 {
			continue
		}
		if !delimiters[0].closed {
			return nil, fmt.Errorf("unterminated quote in heredoc delimiter %q", rest)
		}
		docs = append(docs, dockerHeredoc{name: delimiters[0].value, chomp: chomp})
	}
	return docs, nil
}

// Find the wrapped instruction and its argument view; an incomplete header cannot introduce a body here.
// We preserve the old empty-result behavior and leave instruction grammar to Docker.
func dockerOnbuildInstruction(args string, escape rune) (string, string) {
	end := strings.IndexFunc(args, unicode.IsSpace)
	if end < 0 {
		return "", ""
	}
	return args[:end], dockerArgsAfterFlags(args[end:], escape)
}

// We keep both spellings: raw text locates the word, and its unquoted value matches a heredoc delimiter.
type dockerWord struct {
	raw    string // original word, including quotes and escapes
	value  string // word after removing quotes and applicable escapes
	start  int    // byte offset in the header inspection copy
	closed bool   // every opened quote in this word has closed
}

// Skip leading flag words only to distinguish JSON and ONBUILD headers when
// locating heredocs. This does not validate flags or remove them from tokens.
func dockerArgsAfterFlags(args string, escape rune) string {
	args = strings.TrimSpace(args)
	if !strings.HasPrefix(args, "--") {
		return args
	}
	for _, word := range dockerWords(args, escape, false) {
		if !strings.HasPrefix(word.raw, "--") {
			return args[word.start:]
		}
		if word.value == "--" {
			return strings.TrimSpace(args[word.start+len(word.raw):])
		}
	}
	return ""
}

// Split a header into words while retaining raw spans and quote-removed values.
// The latter are used only for literal heredoc delimiters, never interpolation.
// Whitespace immediately after << is part of its delimiter word in BuildKit.
func dockerWords(text string, escape rune, heredoc bool) []dockerWord {
	var words []dockerWord
	for current := 0; current < len(text); {
		r, width := utf8.DecodeRuneInString(text[current:])
		if unicode.IsSpace(r) {
			current += width
			continue
		}
		start := current
		var value strings.Builder
		var quote rune
		for current < len(text) {
			r, width = utf8.DecodeRuneInString(text[current:])
			if quote == 0 && unicode.IsSpace(r) {
				break
			}
			current += width
			if quote != 0 && r == quote {
				quote = 0
				continue
			}
			if quote == 0 && (r == '\'' || r == '"') {
				quote = r
				continue
			}
			if r == escape && quote != '\'' && current == len(text) {
				continue
			}
			// Quotes decide which escapes consume the next rune; keep that rule in one helper.
			if size := dockerEscapeWidth(text, current, r, quote, escape); size > 0 {
				next, _ := utf8.DecodeRuneInString(text[current:])
				value.WriteRune(next)
				current += size
				continue
			}
			value.WriteRune(r)
			// Only an unquoted << keeps the following whitespace inside this word.
			if !heredoc || quote != 0 || r != '<' || current == len(text) || text[current] != '<' {
				continue
			}
			value.WriteByte('<')
			current++
			space, end := dockerHeredocSpace(text, current)
			value.WriteString(space)
			current = end
		}
		words = append(words, dockerWord{
			raw: text[start:current], value: value.String(), start: start, closed: quote == 0,
		})
	}
	return words
}

// Return the width of the rune consumed by an escape, or zero when the escape stays literal.
// Single quotes disable escapes; double quotes allow only quote, dollar, and the escape character.
func dockerEscapeWidth(text string, current int, char, quote, escape rune) int {
	if char != escape || quote == '\'' || current == len(text) {
		return 0
	}
	next, size := utf8.DecodeRuneInString(text[current:])
	if quote == 0 || next == '"' || next == '$' || next == escape {
		return size
	}
	return 0
}

// Consume the whitespace after << in the header copy, stopping before its delimiter word.
// This keeps the existing heredoc word boundary without adding shell interpretation.
func dockerHeredocSpace(text string, current int) (string, int) {
	var space strings.Builder
	for current < len(text) {
		next, size := utf8.DecodeRuneInString(text[current:])
		if !unicode.IsSpace(next) {
			break
		}
		space.WriteRune(next)
		current += size
	}
	return space.String(), current
}

// Accumulates alphanumeric chars into text buffer, then looks up in DocklettTokenKeywords map
func (s *Scanner) scanDocklettToken() (tokenType token.TokenType, literal any, error error) {
	nameStart := s.current
	for !s.isAtEnd() { // read until space or non-letter/digit
		nextChar := s.peekChar()
		if !unicode.IsLetter(nextChar) && !unicode.IsDigit(nextChar) {
			break
		}
		s.advanceChar()
	}
	text := s.Source[nameStart:s.current]
	docklettTokenType, found := token.DocklettTokenKeywords[text]
	if found {
		return docklettTokenType, nil, nil
	}
	return token.ILLEGAL, nil, s.scanError(fmt.Sprintf("unexpected Docklett token: %q", "@"+text))
}

// Accumulates alphanumeric chars, checks Docklett keywords first if flag set,
// then Docker keywords (queues DOCKER_ARGS as pending), else returns identifier.
// args used to be queued here; ScanSource now reads them only at an instruction boundary.
func (s *Scanner) scanKeywordsAndIdentifierTokens() (tokenType token.TokenType, literal any, error error) {
	for !s.isAtEnd() { // read until space or non-letter/digit
		nextChar := s.peekChar()
		if !unicode.IsLetter(nextChar) && !unicode.IsDigit(nextChar) {
			break
		}
		s.advanceChar()
	}
	text := s.Source[s.start:s.current]
	// if our lexeme starts with a @ we are using Docklett, prioritize Docklett keywords
	// The lookup itself does not change mode; only Docklett mode can use these keyword types.
	docklettTokenType, docklettFound := token.DocklettTokenKeywords[text]
	if s.docklett && docklettFound {
		return docklettTokenType, nil, nil
	}

	// if this is a Docker keyword, emit DOCKER_KEYWORD now and queue DOCKER_ARGS for next cycle
	// the keyword is still emitted here, but ScanSource now handles args in the same cycle.
	// inside Docklett expressions, keep the reserved keyword and leave the remaining tokens for the parser.
	_, found := token.DockerTokenKeywords[strings.ToUpper(text)]
	if !found {
		return token.IDENTIFIER, text, nil
	}
	// A vanilla keyword needs a separator, but a reserved word in an expression owns no Docker args.
	if s.atLineStart && !s.docklett && !s.isAtEnd() && !unicode.IsSpace(s.peekChar()) {
		return token.ILLEGAL, nil, s.scanError("expected whitespace after Docker instruction keyword")
	}
	return token.DOCKER_KEYWORD, nil, nil
}
