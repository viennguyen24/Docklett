# Docklett Philosophy

Software and system philosophies for how we design, implement, and maintain this project. For architecture see [design.md](design.md). For vocabulary see [glossary.md](glossary.md).

## What Docklett is (and is not)

Docklett adds custom compile-time constructs (`@SET`, `@IF`, `@FOR`, …) on top of Dockerfile. It does not replace Docker or BuildKit, and it does not redefine original Dockerfile commands like `FROM` / `RUN` / `COPY`.

The input file is still a `Dockerfile` — same name, same extension. Extra lines are just syntax we own. We do not ship a second build CLI that CI is supposed to adopt.

## Put a `# syntax=` line at the top and, if you need them, `@` directives in the same file. Existing `docker build`, `docker buildx`, or bake jobs stay as they are.

```dockerfile
# syntax=yourname/docklett:1.0
FROM ubuntu:22.04
@SET MODE = "prod"
...
```

BuildKit loads the Docklett frontend from that `# syntax=` image the same way it loads `docker/dockerfile`.

The local `docklett` binary is for us: run the compiler, dump tokens, debug the rewrite. It is not how end users or pipelines should build images. Adding a custom build command to every workflow is exactly the friction this project is trying to avoid.

One precision note: “drop-in” here means same CLI and same pipelines. It does not mean the stock Dockerfile frontend understands `@SET`. Without `# syntax=…/docklett`, those lines are invalid Dockerfile. A file with no `@` lines is still a normal Dockerfile.

## What we process vs what we pass through

Our compiler only owns Docklett syntax. Vanilla instructions keep Docker’s meaning. After rewrite, what remains must be something `dockerfile2llb` already understands.

We call `dockerfile2llb` for the Dockerfile → LLB step. We do not reimplement `RUN`/`COPY`/… as LLB ops.

If the input has no `@` directives and is routed through Docklett, the rewrite should be an identity (same effective Dockerfile).

## Compiler work vs engine work

| Docklett | BuildKit / Docker |
|----------|-------------------|
| Lex, parse, AST | `dockerfile2llb` |
| Evaluate `@SET` / `@IF` / `@FOR` at compile time | LLB, solver, cache |
| Substitute `${…}` from Docklett’s environment | Workers, image export |
| Report Docklett errors before offload | Failures during solve / export |

Once every Docklett-only construct is gone, Docklett is done. The thing users get from `docker build` is still a normal image.

The front half of the codebase can keep Crafting Interpreters shapes (scanner, recursive-descent parser, AST, visitor). The back half must stay a translator to Dockerfile text/AST — not a tree-walk interpreter that “runs” the build, and not a custom LLB builder for each Docker instruction.

`interpreter/` is leftover from before the pivot. Do not add product features there. The path we want is translator → vanilla Dockerfile → `dockerfile2llb`.

## Code hygiene

- Smallest change that keeps the compiler/engine split intact.
- One product path: `# syntax=` frontend, rewrite, then `dockerfile2llb`. Do not maintain a second backend (interpreter *and* translator, or direct LLB *and* Dockerfile rewrite).
- Docklett bugs surface as compile errors. Build failures after offload belong to BuildKit.
- Match the style already in the tree. No cleanup refactors unless someone asked for them.
- When docs and code disagree, fix the docs or the code — do not leave aspirational DESIGN prose as if it were implemented.

## Testing

Almost none today. `scanner_test.go` prints tokens; it does not assert much. Do not stand up a test framework unless asked. If we add tests later, test our pipeline (tokens, AST, rewritten Dockerfile). Do not re-test BuildKit.
