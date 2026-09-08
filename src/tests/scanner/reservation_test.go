package scanner_test

import (
	"testing"

	"docklett/compiler/token"
	"docklett/tests/testutil"
)

func TestDockerNameReservation(t *testing.T) {
	cases := []testutil.Case{
		{
			Name:   "run_uppercase_reserved_not_docker_args",
			Source: "@SET RUN = 1\n",
			Want: []testutil.TokenWant{
				testutil.Tok(token.SET, "@SET", nil),
				testutil.Tok(token.DOCKER_KEYWORD, "RUN", nil),
				testutil.Tok(token.ASSIGN, "=", nil),
				testutil.Tok(token.NUMBER, "1", 1),
				testutil.Tok(token.NLINE, "\n", nil),
				testutil.Tok(token.EOF, "", nil),
			},
		},
		{
			Name:   "run_lowercase_reserved_not_docker_args",
			Source: "@SET run = 1\n",
			Want: []testutil.TokenWant{
				testutil.Tok(token.SET, "@SET", nil),
				testutil.Tok(token.DOCKER_KEYWORD, "run", nil),
				testutil.Tok(token.ASSIGN, "=", nil),
				testutil.Tok(token.NUMBER, "1", 1),
				testutil.Tok(token.NLINE, "\n", nil),
				testutil.Tok(token.EOF, "", nil),
			},
		},
		{
			Name:   "run_mixed_case_reserved_not_docker_args",
			Source: "@SET Run = 1\n",
			Want: []testutil.TokenWant{
				testutil.Tok(token.SET, "@SET", nil),
				testutil.Tok(token.DOCKER_KEYWORD, "Run", nil),
				testutil.Tok(token.ASSIGN, "=", nil),
				testutil.Tok(token.NUMBER, "1", 1),
				testutil.Tok(token.NLINE, "\n", nil),
				testutil.Tok(token.EOF, "", nil),
			},
		},
		{
			Name:   "reservation_does_not_inspect_string_contents",
			Source: "@SET x = \"RUN @END\"\n",
			Want: []testutil.TokenWant{
				testutil.Tok(token.SET, "@SET", nil),
				testutil.Tok(token.IDENTIFIER, "x", "x"),
				testutil.Tok(token.ASSIGN, "=", nil),
				testutil.Tok(token.STRING, "\"RUN @END\"", "RUN @END"),
				testutil.Tok(token.NLINE, "\n", nil),
				testutil.Tok(token.EOF, "", nil),
			},
		},
	}

	for _, c := range cases {
		testutil.RunCase(t, c)
	}
}
