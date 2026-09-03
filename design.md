# Docklett Design

Architecture of the code and of the product we are aiming at. Rules: [philosophy.md](philosophy.md). Vocabulary: [glossary.md](glossary.md). Grammar sketch: [design/grammar.txt](design/grammar.txt).

This file replaces the “compile straight to LLB” story in [design/DESIGN.md](design/DESIGN.md).

## 1. Tech stack

| Layer | Choice |
|-------|--------|
| Language | Go 1.25.5+ (`src/go.mod`, module `docklett`) |
| Dependencies today | None. Stdlib only; BuildKit is not in `go.mod` yet |
| CLI | stdlib `flag` (`-file` / `-F`) |
| Product deps (planned) | BuildKit `frontend/gateway`, `frontend/dockerfile/dockerfile2llb` |
| Tests today | Scanner smoke prints. Not a real suite |

## 2. Layout

```
Docklett/
├── AGENT.md / CLAUDE.md / AGENTS.md   # onboarding (symlinks)
├── philosophy.md, design.md, glossary.md
├── README.md
├── design/
│   ├── DESIGN.md      # old; superseded by this file
│   ├── grammar.txt
│   └── sources.txt    # BuildKit links
├── scripts/           # hook installers
├── .githooks/         # go fmt + go vet
└── src/
    ├── main.go        # → CLI → compiler.Run
    ├── go.mod
    ├── cli/           # harness flags
    └── compiler/
        ├── compiler.go
        ├── token/
        ├── scanner/
        ├── parser/
        ├── ast/
        ├── translator/   # product backend (incomplete; still has LLB stubs)
        ├── interpreter/  # legacy
        ├── error/
        └── util/
```

## 3. Execution flow

### Product path (target)

```
docker build .
  → BuildKit sees "# syntax=…/docklett"
  → Docklett frontend
       → Scanner → Parser → AST
       → Translator (@SET / @IF / @FOR at compile time, ${…})
       → vanilla Dockerfile
       → dockerfile2llb.Dockerfile2LLB(…)
  → LLB → Solver → Image
```

Cutover rule: as soon as the rewrite contains only vanilla Dockerfile, call `dockerfile2llb`. Do not turn `RUN`/`COPY`/… into LLB ops in our code.

| Stage | Who |
|-------|-----|
| Scan, parse, AST, rewrite | Docklett |
| Dockerfile → LLB | `dockerfile2llb` |
| Solve, cache, workers, export | buildkitd |

Two different “outputs”:

1. Compiler result (internal): vanilla Dockerfile. Users usually never see a file on disk.
2. Product result: the image from `docker build`.

### What the code does today

```
docklett -file <path>
  → cli.ParseArgs
  → compiler.Run
       → Scanner.ReadSource + ScanSource
       → keep tokens
       → stop
```

Parser, translator, and any BuildKit frontend are not hooked into `Compiler.Run`. The CLI exists so we can drive the compiler while developing. It is not the integration story for CI.

## 4. User integration

Change the Dockerfile. Leave the pipeline alone.

```dockerfile
# syntax=yourname/docklett:1.0
FROM ubuntu:22.04
@SET MODE = "prod"
@IF MODE == "prod"
RUN apt-get install -y ssl-cert
@ELSE
RUN apt-get install -y vim
@END
WORKDIR /app
COPY . .
```

```bash
docker build .
```

Cost of adoption: publish a Docklett frontend image, and add one `# syntax=` line when you use `@` features. Filename stays `Dockerfile`.

If someone puts `@SET` in a file without `# syntax=…/docklett`, the stock Dockerfile frontend will reject it. That is expected.

Local compiler check:

```bash
cd src && go run main.go -file path/to/Dockerfile
```

Useful for us. Wrong thing to put in a GitHub Action as “the way to build.”

### BuildKit code we expect to import

| Import | Why |
|--------|-----|
| `frontend/gateway` + a published `# syntax=` image | How builds actually invoke Docklett |
| `dockerfile2llb` after rewrite | Official Dockerfile → LLB |

| Do not import for product path | Why |
|--------------------------------|-----|
| Hand-built LLB via `client/llb` per Docker instruction | Reimplements dockerfile2llb |
| CLI calling `client.Solve` as the main UX | New build command; CI friction |
| Solver / worker / cache packages as our logic | Engine internals |

## 5. Components

### CLI (`src/cli`)

Reads `-file` / `-F` or a positional path, checks the file exists, hands the path to `compiler.Run`. Product builds do not go through this binary.

### Compiler (`src/compiler/compiler.go`)

Pipeline driver. Today: scan only. Target inside the frontend: scan → parse → translate, then `dockerfile2llb`.

### Token (`src/compiler/token`)

`Token`, `TokenType`, `DockerTokenKeywords`, `DocklettTokenKeywords`. A Docker verb becomes `DOCKER_KEYWORD` plus a `DOCKER_ARGS` token for the rest of the logical line. Docklett keywords are their own types (`SET`, `IF`, …).

### Scanner (`src/compiler/scanner`)

One pass over the source.

On a normal Docker line it matches an instruction keyword, then eats the rest of the logical line as `DOCKER_ARGS` (backslash continuations included). A leading `@` puts that line into Docklett mode (keywords, identifiers, literals, operators). Emits `NLINE` and `EOF`. Illegal input → `ScanError`.

### Parser (`src/compiler/parser`)

Recursive descent. Expression precedence follows the Crafting Interpreters layout.

It accepts `@SET`, docker statements, `@IF` / `@ELIF` / `@ELSE` / `@END`, `@FOR … IN … @END`, and expression statements. Output is `[]ast.Statement`. On syntax errors it uses panic-mode `synchronize()` and collects `ParseError`s (joined).

The package is written. `Compiler.Run` does not call it yet.

### AST (`src/compiler/ast`)

Expressions: literal, variable, unary, binary, logical, grouping, assignment, array literal, range.

Statements: expression statement, `@SET` declaration, block, if, for, `DockerStatement` (keyword + raw args).

Visitors for both expression and statement trees.

### Translator (`src/compiler/translator`) — intended backend

What it should do:

1. Walk the AST with a compile-time `Environment`.
2. `@SET` / assignment update that environment.
3. `@IF` / `@ELIF` / `@ELSE`: evaluate the condition; keep one branch.
4. `@FOR`: evaluate the iterable (array or `range`); unroll the body (there is a max-iteration guard in the code).
5. On `DockerStatement`, substitute `${name}` from the environment and emit a vanilla instruction.
6. When nothing Docklett-specific remains, pass the result to `dockerfile2llb`.

What it does now: some control-flow scaffolding. Several methods still look like they were going to emit LLB. Treat that as pivot residue. The emit we want is Dockerfile, not hand-built ops.

### Interpreter (`src/compiler/interpreter`) — legacy

Tree-walk evaluator from the first approach. Docker and `For` handling are incomplete. Do not put product features here.

### Errors (`src/compiler/error`)

| Type | Phase |
|------|--------|
| `ScanError` | Lexing |
| `ParseError` | Parsing (sync + collect) |
| `TranslatorError` | Rewrite / compile-time eval |
| Interpreter “runtime” errors | Legacy only |

Docklett failures are compile-time. Solve/export failures are BuildKit’s. Comments that say “the compiler builds LLB” are wrong under the current plan; the compiler builds vanilla Dockerfile, then offloads.

## 6. Dockerfile and BuildKit (why we sit where we sit)

A Dockerfile is an input format. The stock BuildKit path looks like this:

1. Client asks buildkitd to build.
2. A frontend (usually `docker/dockerfile`, chosen by `# syntax=`) parses the file.
3. That frontend lowers instructions to LLB.
4. The solver runs the graph.
5. Workers execute ops; the client gets an image/ref.

Docklett replaces step 2’s frontend. We rewrite `@…` away, then reuse upstream step 3 via `dockerfile2llb` instead of rewriting it.

### Upstream docs worth reading

| Link | For |
|------|-----|
| [docs/dev/README.md](https://github.com/moby/buildkit/blob/master/docs/dev/README.md) | LLB, Frontend, Solver, Vertex, Op, Edge, Result, Worker |
| [docs/dev/request-lifecycle.md](https://github.com/moby/buildkit/blob/master/docs/dev/request-lifecycle.md) | End-to-end request path |
| [docs/dev/dockerfile-llb.md](https://github.com/moby/buildkit/blob/master/docs/dev/dockerfile-llb.md) | Dockerfile → LLB |
| [docs/dev/solver.md](https://github.com/moby/buildkit/blob/master/docs/dev/solver.md) | Solve graph and cache |
| [control.proto](https://github.com/moby/buildkit/blob/master/api/services/control/control.proto) | buildkitd Control API |
| [gateway.proto](https://github.com/moby/buildkit/blob/master/frontend/gateway/pb/gateway.proto) | External frontend ↔ daemon |
| [ops.proto](https://github.com/moby/buildkit/blob/master/solver/pb/ops.proto) | LLB wire format |
| [client/llb/](https://github.com/moby/buildkit/tree/master/client/llb) | LLB Go builders (literacy only for us) |
| dockerfile2llb in BuildKit | The offload we actually want |

Shorter dump: [design/sources.txt](design/sources.txt).

## 7. Syntax

### Parser directives

These are special comments BuildKit understands, not instructions:

| Directive | Role |
|-----------|------|
| `# syntax=image:tag` | Which frontend image parses the file. Our product entry. |
| `# escape=` | Dockerfile escape character |

### Vanilla instructions

Docker’s instruction set. After we tokenize them, we pass them through:

| Instruction | Summary |
|-------------|---------|
| `FROM` | Base image / stage |
| `RUN` | Run commands during the build |
| `CMD` | Default container command |
| `LABEL` | Image metadata |
| `EXPOSE` | Declared ports |
| `ENV` | Environment variables |
| `ADD` | Copy with archive/URL behavior |
| `COPY` | Copy from context or stages |
| `ENTRYPOINT` | Container entrypoint |
| `VOLUME` | Volume mount points |
| `USER` | User/group for later steps |
| `WORKDIR` | Working directory |
| `ARG` | Build-time variables |
| `ONBUILD` | Trigger instructions for downstream builds |
| `STOPSIGNAL` | Default stop signal |
| `HEALTHCHECK` | Health check config |
| `SHELL` | Shell for shell-form commands |
| `MAINTAINER` | Deprecated author field |

`DockerTokenKeywords` in `token.go` lists all of the above, including `MAINTAINER`. Each becomes `DOCKER_KEYWORD` + `DOCKER_ARGS` (opaque remainder of the logical line). We do not yet parse every modern flag or heredoc form; that detail is left for `dockerfile2llb`.

### Docklett syntax

`@` starts a directive. Keyword lexemes in code are uppercase except `range`:

| Construct | Form | What it does |
|-----------|------|----------------|
| Variable | `@SET name` or `@SET name = expr` | Compile-time binding |
| Conditional | `@IF expr` … `@ELIF expr` … `@ELSE` … `@END` | Keep one branch at compile time |
| Loop | `@FOR name IN expr` … `@END` | Unroll at compile time |
| Booleans | `TRUE` / `FALSE` | Literals in Docklett mode |
| Range | `range(start, end)` or with `step` | Iterable for `@FOR` |
| Array | `[expr, …]` | Iterable literal |
| Interpolation | `${name}` inside Docker args | Filled in by the translator |

Operators the scanner knows: `=` `==` `!=` `+` `-` `*` `/` `+=` `-=` `*=` `/=` `<` `<=` `>` `>=` `!` `&&`, and `||` (`OR` in the type set), plus `()` and `[]`.

### Where docs disagree with code

| Topic | Trust this | Ignore this |
|-------|------------|-------------|
| Directive case | `@SET`, `@IF`, … | Old DESIGN’s `@if` |
| Interpolation | `${name}` | Old DESIGN’s `{{name}}` |
| Boolean connectives | Scanner: `&&` / `||` | `grammar.txt`: `"and"` / `"or"` |
| Backend | Translator → Dockerfile → dockerfile2llb | Old DESIGN: Interpreter → LLB Builder |

When unsure, read `token.go` and this file, not `design/DESIGN.md` examples.

## 8. Maturity

| Area | Status |
|------|--------|
| Scanner | Works |
| Parser / AST | Written, not called from `Run` |
| Translator | Scaffold; LLB-shaped leftovers |
| Interpreter | Partial legacy |
| BuildKit frontend / dockerfile2llb | Not started |
| `# syntax=` image | Not started |
| Tests | Scanner prints only |
