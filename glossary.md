# Docklett Glossary

Use these terms when talking about the project so “frontend,” “runtime,” and “compiler output” mean the same thing to everyone.

## Compiler

| Term | Meaning here |
|------|----------------|
| Source language | Input: Dockerfile instructions plus `@` directives and Docklett expressions. |
| Target language | Vanilla Dockerfile after Docklett constructs are removed or expanded. |
| Lexeme | Character span matched in the source (`@SET`, `FROM`, `==`, …). |
| Token | Lexeme + type + position (`token.Token`). What the parser reads. |
| Scanner / lexer | Source → tokens. Switches between Docker-line mode and `@` Docklett mode. |
| Parser | Tokens → AST (or parse errors). Recursive descent. |
| AST | Statement/expression tree. Structure without incidental punctuation. |
| Visitor | AST nodes expose `Accept` so a translator walks the tree by type. |
| Environment | Compile-time name → value scopes for `@SET`, conditions, and loops. |
| Compile-time | Docklett’s work before BuildKit runs: scan, parse, rewrite, expression eval. |
| Runtime (engine) | BuildKit execution after offload (solve, run ops, produce an image). Not the old tree-walk interpreter. |
| Translation / lowering | AST → vanilla Dockerfile (branches chosen, loops unrolled, vars substituted). |
| Directive | Docklett form that starts with `@` (`@IF`, `@FOR`, `@SET`, …). |
| Pass-through instruction | Vanilla Dockerfile line we keep for `dockerfile2llb` without changing its meaning. |
| Unrolling | Turning one `@FOR` body into N copies of its inner instructions at compile time. |
| Interpolation | Replacing `${name}` in Docker args with compile-time values. |

## Product packaging

| Term | Meaning here |
|------|----------------|
| `# syntax=` | BuildKit comment that names the frontend image. This is how users opt into Docklett. |
| Frontend (BuildKit) | Process that turns build input into LLB. Stock image is `docker/dockerfile`; we ship our own. |
| CLI harness | Local `docklett` binary for compiler debugging. Not the build entry users run in CI. |
| Identity compile | Rewrite of a Dockerfile that had no `@` lines into an equivalent vanilla Dockerfile. |

## BuildKit

| Term | Meaning here |
|------|----------------|
| LLB | Low-Level Build: BuildKit’s DAG of build ops. We do not hand-author one Op per Docker instruction. |
| dockerfile2llb | Upstream Dockerfile → LLB conversion. What we call after rewrite. |
| Op | LLB operation kind (exec, source, file, …) in BuildKit’s ops protobuf. |
| Vertex | Node in the solve graph for one op instance. |
| Edge | Dependency between vertices. |
| Solver | Evaluates the LLB/solve graph (scheduling, cache). |
| Gateway | API an external frontend container uses to talk to buildkitd and submit LLB. |
| Worker | Executes ops (containers, snapshots, etc.). |
| Result | Solve output returned to the client (image/ref, …). |

## Names in this repo

| Term | Meaning here |
|------|----------------|
| Docklett | This project: Dockerfile-extending compiler and (eventually) a BuildKit frontend image. |
| DockerStatement | AST node for one vanilla instruction: keyword + raw args string. |
| Translator | Package meant to evaluate Docklett control flow and emit vanilla Dockerfile. |
| Interpreter | Old Crafting Interpreters evaluator. Not the product path. |
