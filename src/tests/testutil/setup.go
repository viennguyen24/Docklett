/*
Package testutil is the shared Go test helpers for Docklett compiler suites.

Use this package from any suite under src/tests/<component>/. It is not a
mock framework and not a second product entry point. It only builds fixtures,
runs the public scan seam, and compares observable results.

Typical success case:

	cases := []testutil.Case{{
		Name:   "keyword_and_args",
		Source: "RUN echo hi\n",
		Want: []testutil.TokenWant{
			testutil.Tok(token.DOCKER_KEYWORD, "RUN", nil),
			testutil.Tok(token.DOCKER_ARGS, "echo hi", nil),
			testutil.Tok(token.NLINE, "\n", nil),
			testutil.Tok(token.EOF, "", nil),
		},
	}}
	for _, c := range cases {
		testutil.RunCase(t, c)
	}

Typical error case: set WantErr instead of Want.

	testutil.Case{
		Name:    "unknown_directive",
		Source:  "@BOGUS\n",
		WantErr: &testutil.ScanErrExpect{Line: 1},
	}

When you need the live *scanner.Scanner (reuse between scans, inspect Tokens
after failure), call NewScanner and ScanSource yourself, then AssertTokens or
AssertScanError. Do not touch unexported scanner fields from tests.
*/
package testutil

import (
	"errors"
	"reflect"
	"testing"

	compileError "docklett/compiler/error"
	"docklett/compiler/scanner"
	"docklett/compiler/token"
)

// DefaultSourceName is the SourceName used when a case does not set File.
// ScanError.File assertions compare against this unless ScanErrExpect.File is set.
const DefaultSourceName = "test.Dockerfile"

// Case is one table row for the scanner seam: source string in, tokens or ScanError out.
// Leave WantErr nil for a success case and fill Want.
// Leave Want empty (or unused) when WantErr is set; RunCase does not compare tokens on error.
type Case struct {
	Name    string
	Source  string
	Want    []TokenWant
	WantErr *ScanErrExpect
}

// TokenWant is the only token surface we lock in tests: Type, Lexeme, Literal.
// Positions are intentionally omitted so refactors of columns do not churn the suite.
type TokenWant struct {
	Type    token.TokenType
	Lexeme  string
	Literal any
}

// ScanErrExpect is the only ScanError surface we lock: Line and File.
// Exact Message strings are not part of the contract.
type ScanErrExpect struct {
	Line int
	File string // empty means DefaultSourceName
}

// Tok is shorthand for building TokenWant rows in tables.
func Tok(typ token.TokenType, lexeme string, literal any) TokenWant {
	return TokenWant{Type: typ, Lexeme: lexeme, Literal: literal}
}

// NewScanner builds a real scanner fixture with in-memory Source.
// Pass sourceName "" to use DefaultSourceName.
func NewScanner(source, sourceName string) *scanner.Scanner {
	if sourceName == "" {
		sourceName = DefaultSourceName
	}
	return &scanner.Scanner{
		Source:     source,
		SourceName: sourceName,
	}
}

// Scan runs ScanSource on an in-memory source with DefaultSourceName.
// Prefer RunCase for ordinary table tests; use Scan when a helper needs the raw result.
func Scan(source string) ([]token.Token, error) {
	s := NewScanner(source, DefaultSourceName)
	err := s.ScanSource()
	return s.Tokens, err
}

// RunCase executes one Case: scan, then AssertTokens or AssertScanError.
// This is the default entry for scanner table tests.
func RunCase(t *testing.T, c Case) {
	t.Helper()
	t.Run(c.Name, func(t *testing.T) {
		t.Helper()
		got, err := Scan(c.Source)
		if c.WantErr != nil {
			AssertScanError(t, err, *c.WantErr)
			return
		}
		if err != nil {
			t.Fatalf("ScanSource() unexpected error: %v", err)
		}
		AssertTokens(t, got, c.Want)
	})
}

// AssertTokens compares Type, Lexeme, and Literal only, in order.
func AssertTokens(t *testing.T, got []token.Token, want []TokenWant) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("token count: got %d want %d\ngot:  %s\nwant: %s",
			len(got), len(want), formatGot(got), formatWant(want))
	}
	for i := range want {
		g, w := got[i], want[i]
		if g.Type != w.Type || g.Lexeme != w.Lexeme || !literalEqual(g.Literal, w.Literal) {
			t.Errorf("token[%d]: got {%s %q %v} want {%s %q %v}",
				i,
				tokenName(g.Type), g.Lexeme, g.Literal,
				tokenName(w.Type), w.Lexeme, w.Literal,
			)
		}
	}
}

// AssertScanError checks that err is a *ScanError with the expected Line and File.
func AssertScanError(t *testing.T, err error, expect ScanErrExpect) {
	t.Helper()
	if err == nil {
		t.Fatal("expected ScanError, got nil")
	}
	var scanErr *compileError.ScanError
	if !errors.As(err, &scanErr) {
		t.Fatalf("expected *ScanError, got %T: %v", err, err)
	}
	file := expect.File
	if file == "" {
		file = DefaultSourceName
	}
	if scanErr.Line != expect.Line {
		t.Errorf("ScanError.Line: got %d want %d", scanErr.Line, expect.Line)
	}
	if scanErr.File != file {
		t.Errorf("ScanError.File: got %q want %q", scanErr.File, file)
	}
}

func literalEqual(got, want any) bool {
	if got == nil && want == nil {
		return true
	}
	return reflect.DeepEqual(got, want)
}

func tokenName(t token.TokenType) string {
	if name, ok := token.TokenTypeNames[t]; ok {
		return name
	}
	return "UNKNOWN"
}

func formatGot(tokens []token.Token) string {
	out := "["
	for i, tok := range tokens {
		if i > 0 {
			out += " "
		}
		out += "{" + tokenName(tok.Type) + " " + quote(tok.Lexeme) + "}"
	}
	return out + "]"
}

func formatWant(tokens []TokenWant) string {
	out := "["
	for i, tok := range tokens {
		if i > 0 {
			out += " "
		}
		out += "{" + tokenName(tok.Type) + " " + quote(tok.Lexeme) + "}"
	}
	return out + "]"
}

func quote(s string) string {
	return "\"" + s + "\""
}
