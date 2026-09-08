package scanner_test

import (
	"testing"

	"docklett/compiler/token"
	"docklett/tests/testutil"
)

func TestMixedStreams(t *testing.T) {
	cases := []testutil.Case{
		{
			Name: "if_run_end_mode_clears_on_newline",
			Source: "@IF TRUE\n" +
				"RUN echo hi\n" +
				"@END\n",
			Want: []testutil.TokenWant{
				testutil.Tok(token.IF, "@IF", nil),
				testutil.Tok(token.TRUE, "TRUE", nil),
				testutil.Tok(token.NLINE, "\n", nil),
				testutil.Tok(token.DOCKER_KEYWORD, "RUN", nil),
				testutil.Tok(token.DOCKER_ARGS, "echo hi", nil),
				testutil.Tok(token.NLINE, "\n", nil),
				testutil.Tok(token.END, "@END", nil),
				testutil.Tok(token.NLINE, "\n", nil),
				testutil.Tok(token.EOF, "", nil),
			},
		},
		{
			Name: "set_then_docker_then_set",
			Source: "@SET x = 1\n" +
				"FROM alpine\n" +
				"@SET y = 2\n",
			Want: []testutil.TokenWant{
				testutil.Tok(token.SET, "@SET", nil),
				testutil.Tok(token.IDENTIFIER, "x", "x"),
				testutil.Tok(token.ASSIGN, "=", nil),
				testutil.Tok(token.NUMBER, "1", 1),
				testutil.Tok(token.NLINE, "\n", nil),
				testutil.Tok(token.DOCKER_KEYWORD, "FROM", nil),
				testutil.Tok(token.DOCKER_ARGS, "alpine", nil),
				testutil.Tok(token.NLINE, "\n", nil),
				testutil.Tok(token.SET, "@SET", nil),
				testutil.Tok(token.IDENTIFIER, "y", "y"),
				testutil.Tok(token.ASSIGN, "=", nil),
				testutil.Tok(token.NUMBER, "2", 2),
				testutil.Tok(token.NLINE, "\n", nil),
				testutil.Tok(token.EOF, "", nil),
			},
		},
		{
			Name: "blank_lines_between_statements",
			Source: "@SET x = 1\n" +
				"\n" +
				"RUN echo x\n",
			Want: []testutil.TokenWant{
				testutil.Tok(token.SET, "@SET", nil),
				testutil.Tok(token.IDENTIFIER, "x", "x"),
				testutil.Tok(token.ASSIGN, "=", nil),
				testutil.Tok(token.NUMBER, "1", 1),
				testutil.Tok(token.NLINE, "\n", nil),
				testutil.Tok(token.NLINE, "\n", nil),
				testutil.Tok(token.DOCKER_KEYWORD, "RUN", nil),
				testutil.Tok(token.DOCKER_ARGS, "echo x", nil),
				testutil.Tok(token.NLINE, "\n", nil),
				testutil.Tok(token.EOF, "", nil),
			},
		},
	}

	for _, c := range cases {
		testutil.RunCase(t, c)
	}
}
