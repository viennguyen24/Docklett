package scanner_test

import (
	"testing"

	"docklett/tests/testutil"
)

func TestScanErrors(t *testing.T) {
	cases := []testutil.Case{
		{
			Name:    "unknown_directive",
			Source:  "@BOGUS x\n",
			WantErr: &testutil.ScanErrExpect{Line: 1},
		},
		{
			Name:    "unexpected_character",
			Source:  "@SET x = 1^\n",
			WantErr: &testutil.ScanErrExpect{Line: 1},
		},
		{
			Name:    "lone_ampersand",
			Source:  "@IF x & y\n",
			WantErr: &testutil.ScanErrExpect{Line: 1},
		},
		{
			Name:    "lone_pipe",
			Source:  "@IF x | y\n",
			WantErr: &testutil.ScanErrExpect{Line: 1},
		},
		{
			Name:    "unterminated_string",
			Source:  "@SET x = \"oops\n",
			WantErr: &testutil.ScanErrExpect{Line: 1},
		},
		{
			Name:    "unterminated_heredoc",
			Source:  "COPY <<EOF /message.txt\nhello\n",
			WantErr: &testutil.ScanErrExpect{Line: 1},
		},
		{
			Name:    "invalid_number",
			Source:  "@SET x = 999999999999999999999999999999\n",
			WantErr: &testutil.ScanErrExpect{Line: 1},
		},
		{
			Name:    "at_outside_directive_entry",
			Source:  "x @SET y = 1\n",
			WantErr: &testutil.ScanErrExpect{Line: 1},
		},
	}

	for _, c := range cases {
		testutil.RunCase(t, c)
	}
}
