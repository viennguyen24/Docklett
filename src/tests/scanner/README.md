# Scanner tests

Behavioral suite for the lexer: in-memory source → `ScanSource` → token list or `ScanError`.

## What this suite locks

- Docker-line vs Docklett-line modes, and mode reset on `NLINE`
- Vanilla instructions as `DOCKER_KEYWORD` + opaque `DOCKER_ARGS` (flags, `@…` text, `${…}`, inline `#` stay in the blob)
- Continuations (`\` / `# escape=`), heredoc payload through closing delimiter
- Docklett directives and expression tokens (`@SET` / `@IF` / `@FOR`, operators including `||`, literals, `range`)
- Global reservation: Docker names inside Docklett expressions emit `DOCKER_KEYWORD` without swallowing the rest as args
- Lexical failures: unknown `@`, unexpected chars, lone `&` / `|`, unterminated string/heredoc, bad numbers
- Scanner reuse: each `ScanSource` resets; failed scans clear `Tokens`

Expectations come from the scanner contract (`tmp/issues.md` + design), not from reverse-engineering a particular binary.

## Files

| File | Covers |
|------|--------|
| `docker_test.go` | Vanilla lines: keyword+args, case folding, empty args, flags, opaque `@` / `${}`, comments vs `#` in blob, EOF |
| `docklett_test.go` | Directive lines and expression lexing: SET/IF/FOR/ELIF/ELSE/END, operators, arrays, strings |
| `mixed_test.go` | Mixed Docker + Docklett streams; blank lines; mode clear across newlines |
| `multiline_test.go` | `\` continuation, `# escape=\``, heredoc body stays in args |
| `errors_test.go` | Hard `ScanError` cases (unknown directive, lone ops, unterminated forms, bad number, stray `@`) |
| `reservation_test.go` | Reserved Docker names in Docklett expressions; reservation does not inspect string contents |
| `reuse_test.go` | Second scan starts clean; failed scan leaves `Tokens` empty |

## How to run

```bash
cd src
go test ./tests/scanner/
```

Shared helpers: `docklett/tests/testutil` (`Case`, `Tok`, `RunCase`, `NewScanner`, …). See package comment in `testutil/setup.go`.
