package scanner_test

import (
	"testing"

	"docklett/compiler/token"
	"docklett/tests/testutil"
)

func TestDockerLines(t *testing.T) {
	cases := []testutil.Case{
		{
			Name:   "keyword_and_opaque_args",
			Source: "RUN echo hello\n",
			Want: []testutil.TokenWant{
				testutil.Tok(token.DOCKER_KEYWORD, "RUN", nil),
				testutil.Tok(token.DOCKER_ARGS, "echo hello", nil),
				testutil.Tok(token.NLINE, "\n", nil),
				testutil.Tok(token.EOF, "", nil),
			},
		},
		{
			Name:   "keyword_case_insensitive",
			Source: "from alpine:3.19\n",
			Want: []testutil.TokenWant{
				testutil.Tok(token.DOCKER_KEYWORD, "from", nil),
				testutil.Tok(token.DOCKER_ARGS, "alpine:3.19", nil),
				testutil.Tok(token.NLINE, "\n", nil),
				testutil.Tok(token.EOF, "", nil),
			},
		},
		{
			Name:   "empty_args",
			Source: "RUN\n",
			Want: []testutil.TokenWant{
				testutil.Tok(token.DOCKER_KEYWORD, "RUN", nil),
				testutil.Tok(token.DOCKER_ARGS, "", nil),
				testutil.Tok(token.NLINE, "\n", nil),
				testutil.Tok(token.EOF, "", nil),
			},
		},
		{
			Name:   "builder_flags_remain_in_args",
			Source: "RUN --mount=type=cache,target=/root/.npm npm install\n",
			Want: []testutil.TokenWant{
				testutil.Tok(token.DOCKER_KEYWORD, "RUN", nil),
				testutil.Tok(token.DOCKER_ARGS, "--mount=type=cache,target=/root/.npm npm install", nil),
				testutil.Tok(token.NLINE, "\n", nil),
				testutil.Tok(token.EOF, "", nil),
			},
		},
		{
			Name:   "directive_shaped_text_stays_opaque",
			Source: "RUN @SET x = 1\n",
			Want: []testutil.TokenWant{
				testutil.Tok(token.DOCKER_KEYWORD, "RUN", nil),
				testutil.Tok(token.DOCKER_ARGS, "@SET x = 1", nil),
				testutil.Tok(token.NLINE, "\n", nil),
				testutil.Tok(token.EOF, "", nil),
			},
		},
		{
			Name:   "package_name_with_at_stays_opaque",
			Source: "RUN npm install @codex/cli\n",
			Want: []testutil.TokenWant{
				testutil.Tok(token.DOCKER_KEYWORD, "RUN", nil),
				testutil.Tok(token.DOCKER_ARGS, "npm install @codex/cli", nil),
				testutil.Tok(token.NLINE, "\n", nil),
				testutil.Tok(token.EOF, "", nil),
			},
		},
		{
			Name:   "quoted_end_marker_stays_opaque",
			Source: "RUN echo \"@END\"\n",
			Want: []testutil.TokenWant{
				testutil.Tok(token.DOCKER_KEYWORD, "RUN", nil),
				testutil.Tok(token.DOCKER_ARGS, "echo \"@END\"", nil),
				testutil.Tok(token.NLINE, "\n", nil),
				testutil.Tok(token.EOF, "", nil),
			},
		},
		{
			Name:   "hash_inside_args_stays_in_blob",
			Source: "RUN echo hello # not a comment\n",
			Want: []testutil.TokenWant{
				testutil.Tok(token.DOCKER_KEYWORD, "RUN", nil),
				testutil.Tok(token.DOCKER_ARGS, "echo hello # not a comment", nil),
				testutil.Tok(token.NLINE, "\n", nil),
				testutil.Tok(token.EOF, "", nil),
			},
		},
		{
			Name:   "full_line_comment_emits_nothing",
			Source: "# only a comment\nRUN echo hi\n",
			Want: []testutil.TokenWant{
				testutil.Tok(token.NLINE, "\n", nil),
				testutil.Tok(token.DOCKER_KEYWORD, "RUN", nil),
				testutil.Tok(token.DOCKER_ARGS, "echo hi", nil),
				testutil.Tok(token.NLINE, "\n", nil),
				testutil.Tok(token.EOF, "", nil),
			},
		},
		{
			Name:   "interpolation_placeholder_stays_in_args",
			Source: "RUN echo ${name}\n",
			Want: []testutil.TokenWant{
				testutil.Tok(token.DOCKER_KEYWORD, "RUN", nil),
				testutil.Tok(token.DOCKER_ARGS, "echo ${name}", nil),
				testutil.Tok(token.NLINE, "\n", nil),
				testutil.Tok(token.EOF, "", nil),
			},
		},
		{
			Name:   "tabs_and_spaces_insignificant_before_keyword",
			Source: "\t  RUN echo hi\n",
			Want: []testutil.TokenWant{
				testutil.Tok(token.DOCKER_KEYWORD, "RUN", nil),
				testutil.Tok(token.DOCKER_ARGS, "echo hi", nil),
				testutil.Tok(token.NLINE, "\n", nil),
				testutil.Tok(token.EOF, "", nil),
			},
		},
		{
			Name:   "eof_without_trailing_newline",
			Source: "WORKDIR /app",
			Want: []testutil.TokenWant{
				testutil.Tok(token.DOCKER_KEYWORD, "WORKDIR", nil),
				testutil.Tok(token.DOCKER_ARGS, "/app", nil),
				testutil.Tok(token.EOF, "", nil),
			},
		},
	}

	for _, c := range cases {
		testutil.RunCase(t, c)
	}
}
