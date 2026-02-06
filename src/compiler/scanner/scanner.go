package scanner

/*
	Scanner reads the input Dockerfile code and produces a list of tokens for the parser to consume.
*/

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"docklett/compiler/token"
	"docklett/compiler/util"
	"unicode"
	"unicode/utf8"
)

type Scanner struct {
	SourcePath string        // filepath of source code
	SourceName string        // filename of source code
	Source     string        // actual source code
	start      int           // first character of current lexeme
	current    int           // current char in source code
	line       int           // current line in source code
	Tokens     []token.Token // list of tokens generated
	docklett   bool          // flag for whether we are using Docklett extensions
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
func (s *Scanner) ScanSource() error {
	if s.line == 0 {
		s.line = 1
	}

	for !s.isAtEnd() {
		s.start = s.current // begin new lexeme
		tokenType, literal, err := s.scanToken()
		if err != nil {
			return err
		}
		if tokenType == token.ILLEGAL {
			continue
		}
		s.addToken(tokenType, literal)
	}

	s.Tokens = append(s.Tokens, token.Token{
		Type: token.EOF,
		Position: token.Position{
			Line: s.line,
			File: s.SourceName,
			Col:  s.current + 1,
		},
	})
	return nil
}

func (s *Scanner) isAtEnd() bool {
	return s.current >= utf8.RuneCountInString(s.Source)
}

func (s *Scanner) addToken(tokenType token.TokenType, literal any) {
	lexeme, _ := util.ReadSubstring(s.Source, s.start, s.current)
	s.Tokens = append(s.Tokens, token.Token{
		Type:   tokenType,
		Lexeme: lexeme,
		Position: token.Position{
			Line: s.line,
			File: s.SourceName,
			Col:  s.start + 1,
		},
		Literal: literal,
	})
}

func (s *Scanner) advanceChar() rune {
	r, _ := util.ReadSingleChar(s.Source, s.current)
	s.current++
	return r
}

func (s *Scanner) nextMatch(expected rune) bool {
	if s.isAtEnd() {
		return false
	}
	if r, _ := util.ReadSingleChar(s.Source, s.current); r != expected {
		return false
	}
	s.current++
	return true
}

func (s *Scanner) scanToken() (tokenType token.TokenType, literal any, err error) {
	lexeme := s.advanceChar()
	switch lexeme {
	// skip whitespace/newlines without emitting tokens
	case ' ', '\t', '\r':
		return token.ILLEGAL, nil, nil
	case '\n':
		s.line++
		s.docklett = false
		return token.NLINE, nil, nil

	case '=':
		if s.nextMatch('=') {
			return token.EQUAL, nil, nil
		}
		return token.ASSIGN, nil, nil
	case '+':
		if s.nextMatch('=') {
			return token.ADD_ASSIGN, nil, nil
		}
		return token.ADD, nil, nil
	case '-':
		if s.nextMatch('=') {
			return token.SUB_ASSIGN, nil, nil
		}
		return token.SUBTRACT, nil, nil
	case '*':
		if s.nextMatch('=') {
			return token.MULTI_ASSIGN, nil, nil
		}
		return token.MULTI, nil, nil
	case '/':
		if s.nextMatch('=') {
			return token.DIV_ASSIGN, nil, nil
		}
		return token.DIVIDE, nil, nil
	case '<':
		if s.nextMatch('=') {
			return token.LTE, nil, nil
		}
		return token.LESS, nil, nil
	case '>':
		if s.nextMatch('=') {
			return token.GTE, nil, nil
		}
		return token.GREATER, nil, nil
	case '!':
		if s.nextMatch('=') {
			return token.UNEQUAL, nil, nil
		}
		return token.NEGATE, nil, nil
	case '&':
		if s.nextMatch('&') {
			return token.AND, nil, nil
		}
		return token.ILLEGAL, nil, fmt.Errorf("unexpected char: &")
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
		s.docklett = true
		return s.scanDocklettToken()
	default:
		if unicode.IsDigit(lexeme) {
			return s.scanNumberToken()
		}
		if unicode.IsLetter(lexeme) {
			return s.scanKeywordsAndIdentifierTokens()
		}
		return token.ILLEGAL, nil, fmt.Errorf("unexpected char: %q", lexeme)
	}
}

// when encounter a #, ignore the entire line because it's a comment
func (s *Scanner) scanComment() (tokenType token.TokenType, literal string, error error) {
	for !s.isAtEnd() {
		nextChar, _ := util.ReadSingleChar(s.Source, s.current)
		if nextChar == '\n' {
			break
		}
		s.advanceChar()
	}
	return token.ILLEGAL, "", nil
}

// when encounter a ", read every single char next until we meet the ending "
func (s *Scanner) scanStringToken() (tokenType token.TokenType, literal string, error error) {
	strLiteral := ""
	for !s.isAtEnd() {
		nextChar, _ := util.ReadSingleChar(s.Source, s.current)
		if nextChar == '"' {
			break
		}
		// read through new lines
		if nextChar == '\n' {
			s.line++
		}
		strLiteral += string(s.advanceChar())
	}
	if s.isAtEnd() {
		return token.ILLEGAL, "", fmt.Errorf("unterminated string literal")
	}
	s.advanceChar() // consume closing "
	return token.STRING, strLiteral, nil

}

func (s *Scanner) scanNumberToken() (tokenType token.TokenType, literal any, error error) {
	isFloat := false
	for !s.isAtEnd() {
		nextChar, _ := util.ReadSingleChar(s.Source, s.current)
		if !unicode.IsDigit(nextChar) {
			break
		}
		s.advanceChar()
	}
	// floating number case
	nextChar, _ := util.ReadSingleChar(s.Source, s.current)
	if nextChar == '.' {
		isFloat = true
		s.advanceChar() // consume the dot
		for !s.isAtEnd() {
			digitChar, _ := util.ReadSingleChar(s.Source, s.current)
			if !unicode.IsDigit(digitChar) {
				break
			}
			s.advanceChar()
		}
	}
	if isFloat {
		floatLiteral, _ := strconv.ParseFloat(s.Source[s.start:s.current], 64)
		return token.NUMBER, floatLiteral, nil
	}
	intLiteral, _ := strconv.Atoi(s.Source[s.start:s.current])
	return token.NUMBER, intLiteral, nil
}

// Reads chars until newline, tracking last non-space char to detect backslash continuations
func (s *Scanner) scanDockerToken() (tokenType token.TokenType, literal any, error error) {
	var lastNonSpace rune
	for !s.isAtEnd() {
		nextChar, _ := util.ReadSingleChar(s.Source, s.current)
		if nextChar == '\n' {
			// Dockerfile instruction may span multiple lines using \
			continued := lastNonSpace == '\\'
			if continued {
				s.advanceChar() // consume newline ONLY for continuation
				s.line++
				lastNonSpace = 0
				continue
			}
			// Leave final newline for main scanner to emit NLINE token
			break
		}
		if nextChar != ' ' && nextChar != '\t' && nextChar != '\r' {
			lastNonSpace = nextChar
		}
		s.advanceChar()
	}
	return token.DLINE, nil, nil
}

// Accumulates alphanumeric chars into text buffer, then looks up in DocklettTokenKeywords map
func (s *Scanner) scanDocklettToken() (tokenType token.TokenType, literal any, error error) {
	text := ""
	for !s.isAtEnd() { // read until space or non-letter/digit
		nextChar, _ := util.ReadSingleChar(s.Source, s.current)
		if !unicode.IsLetter(nextChar) && !unicode.IsDigit(nextChar) {
			break
		}
		text += string(s.advanceChar())
	}
	docklettTokenType, found := token.DocklettTokenKeywords[text]
	firstChar, _ := util.ReadSingleChar(s.Source, s.start)
	if found {
		return docklettTokenType, nil, nil
	}
	return token.ILLEGAL, nil, fmt.Errorf("unexpected Docklett token: %q", string(firstChar)+text)
}

// Accumulates alphanumeric chars, checks Docklett keywords first if flag set, then Docker keywords (delegates to scanDockerToken), else returns identifier
func (s *Scanner) scanKeywordsAndIdentifierTokens() (tokenType token.TokenType, literal any, error error) {
	text := string(s.Source[s.start])
	for !s.isAtEnd() { // read until space or non-letter/digit
		nextChar, _ := util.ReadSingleChar(s.Source, s.current)
		if !unicode.IsLetter(nextChar) && !unicode.IsDigit(nextChar) {
			break
		}
		text += string(s.advanceChar())
	}
	// if our lexeme starts with a @ we are using Docklett, prioritize Docklett keywords
	if s.docklett {
		if docklettTokenType, docklettFound := token.DocklettTokenKeywords[text]; docklettFound {
			return docklettTokenType, nil, nil
		}
	}

	// if this is a Dockerfile raw line, this entire line will be a token
	_, found := token.DockerTokenKeywords[strings.ToUpper(text)]
	if found {
		return s.scanDockerToken()
	}

	return token.IDENTIFIER, text, nil

}
