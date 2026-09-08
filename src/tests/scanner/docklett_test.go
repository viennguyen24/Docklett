package scanner_test

import (
	"testing"

	"docklett/compiler/token"
	"docklett/tests/testutil"
)

func TestDocklettLines(t *testing.T) {
	cases := []testutil.Case{
		{
			Name:   "set_with_initializer",
			Source: "@SET x = 1\n",
			Want: []testutil.TokenWant{
				testutil.Tok(token.SET, "@SET", nil),
				testutil.Tok(token.IDENTIFIER, "x", "x"),
				testutil.Tok(token.ASSIGN, "=", nil),
				testutil.Tok(token.NUMBER, "1", 1),
				testutil.Tok(token.NLINE, "\n", nil),
				testutil.Tok(token.EOF, "", nil),
			},
		},
		{
			Name:   "set_bare_declaration",
			Source: "@SET x\n",
			Want: []testutil.TokenWant{
				testutil.Tok(token.SET, "@SET", nil),
				testutil.Tok(token.IDENTIFIER, "x", "x"),
				testutil.Tok(token.NLINE, "\n", nil),
				testutil.Tok(token.EOF, "", nil),
			},
		},
		{
			Name:   "if_condition_operators",
			Source: "@IF x == 1 && y != 2 || z\n",
			Want: []testutil.TokenWant{
				testutil.Tok(token.IF, "@IF", nil),
				testutil.Tok(token.IDENTIFIER, "x", "x"),
				testutil.Tok(token.EQUAL, "==", nil),
				testutil.Tok(token.NUMBER, "1", 1),
				testutil.Tok(token.AND, "&&", nil),
				testutil.Tok(token.IDENTIFIER, "y", "y"),
				testutil.Tok(token.UNEQUAL, "!=", nil),
				testutil.Tok(token.NUMBER, "2", 2),
				testutil.Tok(token.OR, "||", nil),
				testutil.Tok(token.IDENTIFIER, "z", "z"),
				testutil.Tok(token.NLINE, "\n", nil),
				testutil.Tok(token.EOF, "", nil),
			},
		},
		{
			Name:   "for_in_range",
			Source: "@FOR i IN range(0, 3)\n",
			Want: []testutil.TokenWant{
				testutil.Tok(token.FOR, "@FOR", nil),
				testutil.Tok(token.IDENTIFIER, "i", "i"),
				testutil.Tok(token.IN, "IN", nil),
				testutil.Tok(token.RANGE, "range", nil),
				testutil.Tok(token.LPAREN, "(", nil),
				testutil.Tok(token.NUMBER, "0", 0),
				testutil.Tok(token.COMMA, ",", nil),
				testutil.Tok(token.NUMBER, "3", 3),
				testutil.Tok(token.RPAREN, ")", nil),
				testutil.Tok(token.NLINE, "\n", nil),
				testutil.Tok(token.EOF, "", nil),
			},
		},
		{
			Name:   "elif_else_end",
			Source: "@ELIF TRUE\n@ELSE\n@END\n",
			Want: []testutil.TokenWant{
				testutil.Tok(token.ELIF, "@ELIF", nil),
				testutil.Tok(token.TRUE, "TRUE", nil),
				testutil.Tok(token.NLINE, "\n", nil),
				testutil.Tok(token.ELSE, "@ELSE", nil),
				testutil.Tok(token.NLINE, "\n", nil),
				testutil.Tok(token.END, "@END", nil),
				testutil.Tok(token.NLINE, "\n", nil),
				testutil.Tok(token.EOF, "", nil),
			},
		},
		{
			Name:   "string_and_array_delimiters",
			Source: "@SET xs = [\"a\", 2.5]\n",
			Want: []testutil.TokenWant{
				testutil.Tok(token.SET, "@SET", nil),
				testutil.Tok(token.IDENTIFIER, "xs", "xs"),
				testutil.Tok(token.ASSIGN, "=", nil),
				testutil.Tok(token.LBRACKET, "[", nil),
				testutil.Tok(token.STRING, "\"a\"", "a"),
				testutil.Tok(token.COMMA, ",", nil),
				testutil.Tok(token.NUMBER, "2.5", 2.5),
				testutil.Tok(token.RBRACKET, "]", nil),
				testutil.Tok(token.NLINE, "\n", nil),
				testutil.Tok(token.EOF, "", nil),
			},
		},
		{
			Name:   "false_literal_and_comparisons",
			Source: "@IF n < 10 && n >= 0 && !FALSE\n",
			Want: []testutil.TokenWant{
				testutil.Tok(token.IF, "@IF", nil),
				testutil.Tok(token.IDENTIFIER, "n", "n"),
				testutil.Tok(token.LESS, "<", nil),
				testutil.Tok(token.NUMBER, "10", 10),
				testutil.Tok(token.AND, "&&", nil),
				testutil.Tok(token.IDENTIFIER, "n", "n"),
				testutil.Tok(token.GTE, ">=", nil),
				testutil.Tok(token.NUMBER, "0", 0),
				testutil.Tok(token.AND, "&&", nil),
				testutil.Tok(token.NEGATE, "!", nil),
				testutil.Tok(token.FALSE, "FALSE", nil),
				testutil.Tok(token.NLINE, "\n", nil),
				testutil.Tok(token.EOF, "", nil),
			},
		},
		{
			Name:   "interpolation_inside_string_stays_literal",
			Source: "@SET msg = \"hello ${name}\"\n",
			Want: []testutil.TokenWant{
				testutil.Tok(token.SET, "@SET", nil),
				testutil.Tok(token.IDENTIFIER, "msg", "msg"),
				testutil.Tok(token.ASSIGN, "=", nil),
				testutil.Tok(token.STRING, "\"hello ${name}\"", "hello ${name}"),
				testutil.Tok(token.NLINE, "\n", nil),
				testutil.Tok(token.EOF, "", nil),
			},
		},
	}

	for _, c := range cases {
		testutil.RunCase(t, c)
	}
}
