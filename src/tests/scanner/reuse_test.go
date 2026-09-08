package scanner_test

import (
	"errors"
	"testing"

	compileError "docklett/compiler/error"
	"docklett/compiler/token"
	"docklett/tests/testutil"
)

func TestScannerReuse(t *testing.T) {
	t.Run("second_scan_starts_clean", func(t *testing.T) {
		s := testutil.NewScanner("RUN echo first\n", testutil.DefaultSourceName)
		if err := s.ScanSource(); err != nil {
			t.Fatalf("first ScanSource: %v", err)
		}
		firstCount := len(s.Tokens)
		if firstCount == 0 {
			t.Fatal("first scan produced no tokens")
		}

		s.Source = "RUN echo second\n"
		if err := s.ScanSource(); err != nil {
			t.Fatalf("second ScanSource: %v", err)
		}

		want := []testutil.TokenWant{
			testutil.Tok(token.DOCKER_KEYWORD, "RUN", nil),
			testutil.Tok(token.DOCKER_ARGS, "echo second", nil),
			testutil.Tok(token.NLINE, "\n", nil),
			testutil.Tok(token.EOF, "", nil),
		}
		testutil.AssertTokens(t, s.Tokens, want)
	})

	t.Run("failed_scan_clears_partial_tokens", func(t *testing.T) {
		s := testutil.NewScanner("@SET x = 1\n", testutil.DefaultSourceName)
		if err := s.ScanSource(); err != nil {
			t.Fatalf("setup ScanSource: %v", err)
		}
		if len(s.Tokens) == 0 {
			t.Fatal("setup scan produced no tokens")
		}

		s.Source = "@BOGUS\n"
		err := s.ScanSource()
		if err == nil {
			t.Fatal("expected ScanError for unknown directive")
		}
		var scanErr *compileError.ScanError
		if !errors.As(err, &scanErr) {
			t.Fatalf("expected *ScanError, got %T: %v", err, err)
		}
		if len(s.Tokens) != 0 {
			t.Fatalf("Tokens after failed scan: got %#v want empty", s.Tokens)
		}
	})
}
