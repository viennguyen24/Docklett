package scanner_test

import (
	"testing"

	"docklett/compiler/token"
	"docklett/tests/testutil"
)

func TestMultilineDocker(t *testing.T) {
	cases := []testutil.Case{
		{
			Name: "backslash_continuation_keeps_next_line_in_args",
			Source: "RUN echo hello \\\n" +
				"    @SET x = 1\n",
			Want: []testutil.TokenWant{
				testutil.Tok(token.DOCKER_KEYWORD, "RUN", nil),
				testutil.Tok(token.DOCKER_ARGS, "echo hello \\\n    @SET x = 1", nil),
				testutil.Tok(token.NLINE, "\n", nil),
				testutil.Tok(token.EOF, "", nil),
			},
		},
		{
			Name: "completed_instruction_allows_next_directive",
			Source: "RUN echo hello\n" +
				"@SET x = 1\n",
			Want: []testutil.TokenWant{
				testutil.Tok(token.DOCKER_KEYWORD, "RUN", nil),
				testutil.Tok(token.DOCKER_ARGS, "echo hello", nil),
				testutil.Tok(token.NLINE, "\n", nil),
				testutil.Tok(token.SET, "@SET", nil),
				testutil.Tok(token.IDENTIFIER, "x", "x"),
				testutil.Tok(token.ASSIGN, "=", nil),
				testutil.Tok(token.NUMBER, "1", 1),
				testutil.Tok(token.NLINE, "\n", nil),
				testutil.Tok(token.EOF, "", nil),
			},
		},
		{
			Name: "escape_directive_uses_backtick_continuation",
			Source: "# escape=`\n" +
				"RUN echo hello `\n" +
				"    world\n",
			Want: []testutil.TokenWant{
				testutil.Tok(token.NLINE, "\n", nil),
				testutil.Tok(token.DOCKER_KEYWORD, "RUN", nil),
				testutil.Tok(token.DOCKER_ARGS, "echo hello `\n    world", nil),
				testutil.Tok(token.NLINE, "\n", nil),
				testutil.Tok(token.EOF, "", nil),
			},
		},
		{
			Name: "heredoc_body_and_closer_stay_in_args",
			Source: "COPY <<EOF /message.txt\n" +
				"hello\n" +
				"@SET x = 1\n" +
				"EOF\n",
			Want: []testutil.TokenWant{
				testutil.Tok(token.DOCKER_KEYWORD, "COPY", nil),
				testutil.Tok(token.DOCKER_ARGS, "<<EOF /message.txt\nhello\n@SET x = 1\nEOF", nil),
				testutil.Tok(token.NLINE, "\n", nil),
				testutil.Tok(token.EOF, "", nil),
			},
		},
	}

	for _, c := range cases {
		testutil.RunCase(t, c)
	}
}
