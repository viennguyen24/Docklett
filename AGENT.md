# AGENT.md

Agent onboarding. `CLAUDE.md` and `AGENTS.md` are symlinks to this file.

## Project in one paragraph

Docklett compiles Dockerfiles that may contain `@SET` / `@IF` / `@FOR` (and related expressions) into vanilla Dockerfile semantics, then hands the result to BuildKit through `dockerfile2llb`. End users add `# syntax=…/docklett` and keep using `docker build`. The `docklett` CLI is only a local compiler harness.

## Read order

1. [philosophy.md](philosophy.md) — design rules  
2. [design.md](design.md) — architecture, components, syntax  
3. [glossary.md](glossary.md) — terms  
4. [design/grammar.txt](design/grammar.txt) — grammar sketch  
5. [src/compiler/](src/compiler/)

[design/DESIGN.md](design/DESIGN.md) is old (direct LLB / interpreter). Ignore it unless you are archaeology.

## Rules

- Do not map Dockerfile instructions to LLB ops yourself. Call `dockerfile2llb` after rewrite.
- Do not extend `interpreter/` for product work. Product backend is the translator emitting vanilla Dockerfile.
- Do not invent CI build commands. Product entry is `# syntax=`.
- New language features are `@` directives and their expressions only. Vanilla instruction meaning stays with Docker.
- Small diffs. No drive-by refactors.
- No new test harness unless asked.
- Trust root docs + `token.go` over old DESIGN examples. Real syntax is `@IF` and `${var}`, not `@if` and `{{var}}`.

## Comments and module design

This project focuses on design, decisioning, and declaration of responsibility more than coding. Comments carry the rules. Existing comments in a module win over this file if they conflict; update AGENT.md later, do not “clean up” the code comments to match a prettier doc.

Never delete or rewrite an existing comment. Add new ones. When behavior changes, leave the old line and put the new rule on the next line. Keep the original even if the code it described is gone.

Every module gets a package `/* ... */` header. Split packages: full overview on the main file, shorter header on the others for their slice. Add a missing header when you touch that module, inside the task’s scope.

## Comment guidelines and convention:
- Precise wording, not big wording. Nail down the exact definition using clear descriptive, and simple daily words rather than complicated corporate language
- For design and decision related comments, remove all implementation specific detail to make the design independent and can be understood first before looking at the code.
- If user has any specific humanized writing skill, use it

### Package header

Tab-indent the body. Use blank lines between topics. Do not polish into an essay.

Order:

1. What the module does, and the terms it owns (what comes in, what goes out).
2. The job in plain words: which rules run, in what order, under what mode/context; what must be kept in a token; what must be rejected. Say what this module does not decide.
3. If more than one rule set shares the input, number them `1.` / `2.`. Give each a `Shape:`, what counts as payload, and what tokens (or errors) come out.
4. Mode and boundary intuition: when each rule set owns the next character, what flips mode, what clears it, what must not be treated as a new statement start.
5. Hard non-goals at the end (“we do not …”) so the job does not quietly grow.

### Voice

Informal. Developer explaining the walk to another developer. Fragments and rough grammar are fine when the meaning is clear. Prefer “we” for choices (what we emit, refuse, leave for a later phase).

Name the real domain words (lexeme, token, instruction, args, mode, delimiter). Define a term where it first matters. Prefer a `Shape:`, a tiny example, or a `->` flow over abstract prose when the processing order is the point.

No templates like `Responsibility:` / `Invariant:` / `Note:`. No decorative banners. No bold callouts inside comments.

### `//` comments

Above a type: what it holds and what it is not (machine/walk state vs finished values handed to the next phase).

Above a function: how this step fits the design — order, which mode owns the text, failure that matters. One line for a tiny helper. Several lines for a central path when each line adds a real rule.

Inside a function, next to the decision: why this branch is safe or required, not a restatement of the Go.

Struct fields: short inline `//` for meaning and units (one-based, byte vs rune, and so on). A `//` line above a related group is fine when they share one rule.

### Do not

- Rewrite history to sound nicer.
- Dump the whole package design onto every helper.
- Narrate the next line of code when it adds no rule (`i++ // increment`).
- Import a formal “doc comment” style from elsewhere and overwrite this voice.

## Declaration order and control flow

- Put all interface declarations at the top of their Go file, after the header, `package`, and imports, before other type declarations, constants, variables, or functions. Keep each interface's explanatory comments with it.
- Do not introduce nested conditionals or branches (`if` or `switch` inside another branch). Use guard clauses, early returns, and `continue` to handle rejected or completed cases, then write the remaining path in reading order. Split separate decisions into named helpers when needed; their names and comments must keep the governing rule visible.
- These are implementation conventions. Docklett's grammar still determines which nested `@IF` / `@FOR` constructs the compiler accepts.

## Make the rules visible in code

This project prioritizes expressing and enforcing its rules. A reader should be able to follow the code in order and recognize the corresponding language or design rule. Keep checks and actions close to the rule they implement, use the terms from the grammar and glossary, and preserve any ordering the rule requires.

Every optimization or cleanup must include an explanation comment beside the changed code: state the rule, explain how the new implementation still enforces it, and identify any assumptions that make the change valid. A claim such as "faster" or "equivalent" is not enough. If the reader has to reconstruct the rule through clever code or scattered helpers, choose the more direct implementation. The small-diff and no-drive-by-refactor rules still apply. Keep the old comment when replacing behavior; add the correction on the next line.

## What works today

`compiler.Run` scans and stops. Parser and AST exist but are not wired. Translator is incomplete and still has LLB-shaped stubs from the old plan; the target emit is Dockerfile, not hand-built ops. `go.mod` has no BuildKit dependency. Scanner behavioral tests exist under `src/tests/scanner/` with shared helpers in `src/tests/testutil/`.

## Tests

### Layout

Module root is `src/` (`go.mod`). Behavioral suites live under `src/tests/`, not next to production packages.

```
src/tests/
  testutil/     # shared helpers (setup.go) — Case tables, fixtures, assertions
  scanner/      # scanner seam tests + README.md
  # later: parser/, translator/, …
```

Import helpers as `docklett/tests/testutil`. Component suites use an external test package (`scanner_test`, etc.) so they only see exported APIs.

### Suite README (required)

Every folder under `src/tests/<component>/` that holds cases must include a `README.md`. Keep it short. It should cover:

1. Overview — which compiler layer and which public seam
2. Behaviors locked — the rules the suite is protecting
3. Files — what each `*_test.go` is responsible for
4. How to run — the `go test` path for that suite

`testutil/` does not need a suite README; its package comment in `setup.go` is the usage guide. When you add `parser/` or another suite, add its README in the same pass as the first test file.

### How to run

```bash
cd src
go test ./tests/...
go test ./tests/scanner/
go test -v ./tests/scanner/
```

### What to use

- Stdlib `testing` only. No testify, golden frameworks, or a second helper package unless asked.
- Build cases with `testutil.Case`, `testutil.Tok`, and `testutil.RunCase`.
- For checks that need a live `*scanner.Scanner` (reuse, failed-scan cleanup), use `testutil.NewScanner` + `ScanSource`, then `AssertTokens` / `AssertScanError`.
- Reuse compiler types: `token.Token` / `token.TokenType`, `compileError.ScanError`. Do not invent parallel token or error types for tests.
- Do not mock the scanner behind an interface. Fixture = real `Scanner` with in-memory `Source` and `SourceName`.
- Do not drive tests through `ReadSource`, `compiler.Run`, or the CLI. Do not pull in BuildKit / `dockerfile2llb`.
- Read the package comment at the top of `src/tests/testutil/setup.go` for copy-paste usage examples.

### Conventions

- One primary seam per layer. Scanner: source string → `ScanSource` → `[]token.Token` or `*ScanError`.
- Assert observable behavior only. Tokens: `Type`, `Lexeme`, `Literal`. Errors: `*ScanError` plus `Line` and `File` (default file name is `test.Dockerfile`). Do not lock token positions or exact error message prose.
- One behavioral reason to fail per case. One representative input per equivalence class; add a second case only when the rule is different.
- Derive `Want` from the language contract (`tmp/issues.md`, root design docs, `token.go`), not from whatever the current binary emits. If the suite is red, fix the code or deliberately change the contract and the test together.
- Table-driven files grouped by concern. Prefer `testutil.RunCase` for ordinary tables.
- When adding a later component suite, put it at `src/tests/<component>/` with its own `README.md`, extend `testutil` only with helpers that stay layer-agnostic, and keep that component’s public seam as the assertion boundary (e.g. parser: token stream → AST / `ParseError`).

### Scanner contract notes for tests

- Vanilla instruction → `DOCKER_KEYWORD` (verb lexeme only) + opaque `DOCKER_ARGS`. No sub-lexing inside the blob. Flags, `@…` text, `${name}`, and `#` inside args stay in the blob.
- Full-line `#` comments produce no tokens; the following newline still emits `NLINE` when present.
- Continuations and heredoc bodies stay inside that instruction’s `DOCKER_ARGS` through the closer.
- Docklett lines start with `@` + known keywords; lexemes for `@SET` / `@IF` / … retain the `@`. Unknown `@Foo` is a `ScanError`.
- Global reservation: Docker instruction names used inside Docklett expressions emit `DOCKER_KEYWORD` with a nil literal and must **not** swallow the rest of the line as `DOCKER_ARGS`. Parser owns rejecting them as identifiers.
- Successful scans end with `EOF`. Failed scans clear `Tokens`. Each `ScanSource` call resets scanner state.

## Local harness and tools

For now, only read-only tools and edit files inside the codebase is allowed. No need to run / execute

## BuildKit reading

If you touch the frontend or offload boundary, use the link list in [design.md](design.md) section 6 (request lifecycle, dockerfile-llb, gateway, ops.proto, dockerfile2llb).
