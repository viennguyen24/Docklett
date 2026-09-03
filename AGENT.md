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

## What works today

`compiler.Run` scans and stops. Parser and AST exist but are not wired. Translator is incomplete and still has LLB-shaped stubs from the old plan; the target emit is Dockerfile, not hand-built ops. `go.mod` has no BuildKit dependency. No useful tests.

## Local harness and tools

For now, only read-only tools and edit files inside the codebase is allowed. No need to run / execute

## BuildKit reading

If you touch the frontend or offload boundary, use the link list in [design.md](design.md) section 6 (request lifecycle, dockerfile-llb, gateway, ops.proto, dockerfile2llb).
