# Docklett Design Document

## Executive Summary

**Docklett** is a compiler that extends standard Dockerfile syntax with programming language features (variables, conditionals, loops) and compiles them into BuildKit's Low-Level Build (LLB) graph format. The LLB graph is sent directly to Docker's build daemon (BuildKit) for execution, enabling more expressive build definitions while maintaining full compatibility with Docker's build infrastructure.

---

## System Overview

### What Docklett Is

Docklett transforms enhanced Dockerfiles into executable build graphs. Unlike traditional Dockerfile preprocessors that output text-based Dockerfiles, Docklett compiles directly to LLB (Low-Level Build), the intermediate representation used internally by Docker BuildKit.

**Key Differences from Standard Dockerfiles:**
- Supports variables, conditionals (`@if/@else/@end`), and loops (`@for/@end`)
- Compiles to LLB graph (not text output)
- Leverages BuildKit's parallelization and caching
- Type-safe compilation with clear error reporting

### Example Docklett File

```dockerfile
FROM ubuntu:22.04

# Conditional compilation based on MODE variable
@if MODE == "prod"
RUN echo "Running production configuration"
RUN apt-get update && apt-get install -y ssl-cert
@else
RUN echo "Running development configuration"
RUN apt-get update && apt-get install -y vim curl
@end

# Loop expansion for package installation
@for pkg in ["curl", "git", "vim"]
RUN apt-get install -y {{pkg}}
@end

# Traditional Dockerfile commands
WORKDIR /app
COPY . .
RUN make build
```

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                          DOCKLETT COMPILER                          │
│                                                                     │
│  ┌──────────┐   ┌────────┐   ┌─────┐   ┌────────────┐   ┌───────┐ │
│  │ Scanner  │──▶│ Parser │──▶│ AST │──▶│Interpreter │──▶│  LLB  │ │
│  │ (Lexer)  │   │        │   │     │   │(Evaluator) │   │Builder│ │
│  └──────────┘   └────────┘   └─────┘   └────────────┘   └───────┘ │
│       │             │           │             │               │     │
│    Tokens        Syntax      Tree       Semantic         LLB Def   │
│                  Analysis              Analysis                    │
└─────────────────────────────────────────────────────────────────────┘
         │                                                      │
         │ Input: Docklett File                                │
         │ (text)                                              ▼
         │                                           ┌──────────────────┐
         │                                           │  LLB Graph       │
         │                                           │  Definition      │
         │                                           │  (Protobuf/JSON) │
         │                                           └──────────────────┘
         │                                                      │
         │                                                      ▼
         │                                           ┌──────────────────┐
         │                                           │  Docker BuildKit │
         ▼                                           │  Daemon          │
    ┌─────────┐                                     │  (Solver)        │
    │   CLI   │────────────────────────────────────▶│                  │
    └─────────┘         Build Context +              └──────────────────┘
                        LLB Graph                             │
                                                              ▼
                                                    ┌──────────────────┐
                                                    │  Container Image │
                                                    └──────────────────┘
```

---

## Component Definitions

### 1. CLI (Command-Line Interface)

**Purpose**: Entry point for user interaction with Docklett compilation process.

**Responsibilities:**
- Parse command-line arguments (file path, build context, output options)
- Read Docklett source file from disk
- Coordinate compiler pipeline execution
- Handle build context packaging
- Interface with BuildKit daemon via gRPC
- Report compilation errors and build status to user

**Interface:**

```
INPUT:
  - Docklett file path (string)
  - Build context directory (string)
  - Build variables (map[string]string)
  - BuildKit daemon address (string, default: unix:///var/run/buildkit/buildkitd.sock)
  - Output options (image name, tags, registry auth)

OUTPUT:
  - Exit code (0 = success, non-zero = failure)
  - Build logs (stdout/stderr)
  - Image ID or digest (on success)
  - Error messages (on failure)
```

**Commands:**

```bash
docklett build -f Docklett -t myapp:latest .
docklett build --var MODE=prod -f Docklett.prod .
docklett compile -f Docklett -o llb.json  # Debug: output LLB only
docklett validate -f Docklett              # Syntax check only
```

**Error Handling:**
- File I/O errors (file not found, permissions)
- Compilation errors (delegate to compiler components)
- BuildKit connection errors (daemon not running, auth failures)
- Build execution errors (from BuildKit)

---

### 2. Scanner (Lexer)

**Purpose**: Transform raw Docklett source text into a sequence of tokens.

**Responsibilities:**
- Read source file character-by-character
- Recognize lexical patterns (keywords, identifiers, operators, literals)
- Skip whitespace and comments (# for single-line comments)
- Track line and column numbers for error reporting
- Emit tokens with position metadata

**Input:**
```
TYPE: string (Docklett source code)
EXAMPLE:
@if MODE == "prod"
RUN echo "prod"
@end
```

**Output:**
```
TYPE: []Token

Token Structure:
{
  Type:    TokenType  // KEYWORD, IDENTIFIER, STRING, OPERATOR, etc.
  Lexeme:  string     // Actual text from source
  Literal: interface{} // Typed value for literals (strings, numbers)
  Line:    int
  Column:  int
}

EXAMPLE OUTPUT:
[
  {Type: DIRECTIVE_IF,    Lexeme: "@if",   Line: 1, Column: 1},
  {Type: IDENTIFIER,      Lexeme: "MODE",  Line: 1, Column: 5},
  {Type: EQUAL_EQUAL,     Lexeme: "==",    Line: 1, Column: 10},
  {Type: STRING,          Lexeme: "\"prod\"", Literal: "prod", Line: 1, Column: 13},
  {Type: NEWLINE,         Lexeme: "\n",    Line: 1, Column: 19},
  {Type: DOCKER_COMMAND,  Lexeme: "RUN",   Line: 2, Column: 1},
  {Type: DOCKER_ARGS,     Lexeme: "echo \"prod\"", Line: 2, Column: 5},
  {Type: NEWLINE,         Lexeme: "\n",    Line: 2, Column: 16},
  {Type: DIRECTIVE_END,   Lexeme: "@end",  Line: 3, Column: 1},
  {Type: EOF,             Lexeme: "",      Line: 3, Column: 5}
]
```

**Token Types (Enumeration):**

```go
// Docklett Directives
DIRECTIVE_IF, DIRECTIVE_ELSE, DIRECTIVE_END, DIRECTIVE_FOR, DIRECTIVE_IN

// Docker Commands (pass-through)
DOCKER_COMMAND  // FROM, RUN, COPY, WORKDIR, etc.
DOCKER_ARGS     // Everything after the command until newline

// Literals
IDENTIFIER      // Variable names: MODE, pkg
STRING          // "prod", "curl"
NUMBER          // 123, 45.67
BOOLEAN         // true, false
ARRAY_START     // [
ARRAY_END       // ]

// Operators
EQUAL_EQUAL     // ==
NOT_EQUAL       // !=
GREATER         // >
LESS            // <
GREATER_EQUAL   // >=
LESS_EQUAL      // <=
AND             // &&
OR              // ||
NOT             // !

// Template Interpolation
TEMPLATE_START  // {{
TEMPLATE_END    // }}

// Structural
COMMA           // ,
NEWLINE         // \n
EOF             // End of file
```

**Edge Cases:**
- Strings with escaped quotes: `"string with \"quotes\""`
- Multi-line RUN commands ending with backslash: `RUN apt-get update \`
- Comments starting with #: `# This is a comment`
- Template variables in Docker args: `RUN install {{pkg}}`

**Error Examples:**
```
Error: Unterminated string literal
  --> Docklett:5:10
   |
 5 | @if MODE == "prod
   |             ^^^^^ expected closing quote

Error: Invalid character in directive
  --> Docklett:3:1
   |
 3 | @1nvalid
   | ^^^^^^^^ directives must start with letter
```

---

### 3. Parser

**Purpose**: Transform token sequence into Abstract Syntax Tree (AST) representing program structure.

**Responsibilities:**
- Consume tokens from Scanner
- Apply grammar rules via recursive descent parsing
- Build tree structure representing program hierarchy
- Validate syntax correctness
- Report parse errors with context

**Input:**
```
TYPE: []Token (from Scanner)
```

**Output:**
```
TYPE: *AST (Abstract Syntax Tree)

AST Node Types:

// Root node
ProgramNode {
  Statements: []StatementNode
}

// Statement nodes
DockerCommandNode {
  Command:   string        // "FROM", "RUN", "COPY", etc.
  Arguments: []ArgumentNode // Parsed arguments
  Line:      int
}

IfStatementNode {
  Condition:    ExpressionNode
  ThenBranch:   []StatementNode
  ElseBranch:   []StatementNode  // nil if no @else
  Line:         int
}

ForLoopNode {
  Variable:   string             // "pkg"
  Iterable:   ExpressionNode     // ArrayLiteralNode
  Body:       []StatementNode
  Line:       int
}

// Expression nodes
BinaryExpressionNode {
  Left:     ExpressionNode
  Operator: TokenType        // EQUAL_EQUAL, AND, etc.
  Right:    ExpressionNode
}

IdentifierNode {
  Name: string
}

LiteralNode {
  Value: interface{}  // string, int, bool, []interface{}
  Type:  LiteralType  // STRING, NUMBER, BOOLEAN, ARRAY
}

ArrayLiteralNode {
  Elements: []ExpressionNode
}

TemplateLiteralNode {
  Parts: []interface{}  // mix of strings and IdentifierNodes
}
```

**Grammar (BNF-like notation):**

```
program        → statement* EOF

statement      → dockerCommand
               | ifStatement
               | forLoop

dockerCommand  → DOCKER_COMMAND dockerArgs NEWLINE

dockerArgs     → (DOCKER_ARGS | templateLiteral)*

ifStatement    → DIRECTIVE_IF expression NEWLINE
                 statement*
                 (DIRECTIVE_ELSE NEWLINE statement*)?
                 DIRECTIVE_END NEWLINE

forLoop        → DIRECTIVE_FOR IDENTIFIER DIRECTIVE_IN expression NEWLINE
                 statement*
                 DIRECTIVE_END NEWLINE

expression     → equality

equality       → comparison (("==" | "!=") comparison)*

comparison     → term ((">" | ">=" | "<" | "<=") term)*

term           → unary (("&&" | "||") unary)*

unary          → ("!") unary | primary

primary        → IDENTIFIER
               | STRING
               | NUMBER
               | BOOLEAN
               | arrayLiteral
               | "(" expression ")"

arrayLiteral   → "[" (expression ("," expression)*)? "]"

templateLiteral→ TEMPLATE_START IDENTIFIER TEMPLATE_END
```

**Operator Precedence (highest to lowest):**
```
1. Unary:       ! (NOT)
2. Comparison:  ==, !=, <, <=, >, >=
3. Logical AND: &&
4. Logical OR:  ||
```

**Parse Error Examples:**
```
Error: Expected expression after if directive
  --> Docklett:10:4
   |
10 | @if
   |    ^ expected condition expression

Error: Mismatched @end directive
  --> Docklett:15:1
   |
15 | @end
   | ^^^^ no matching @if or @for

Error: Expected 'in' keyword in for loop
  --> Docklett:8:15
   |
 8 | @for pkg of ["curl"]
   |          ^^ expected 'in', found 'of'
```

---

### 4. Interpreter (Semantic Analyzer & Evaluator)

**Purpose**: Walk AST, evaluate control flow, expand templates, and generate linear execution plan.

**Responsibilities:**
- Traverse AST in depth-first order
- Resolve variable references
- Evaluate conditional expressions (if/else)
- Expand for loops (unroll iterations)
- Substitute template variables ({{var}})
- Validate variable scoping and types
- Generate sequence of Docker commands with control flow resolved

**Input:**
```
TYPE: *AST (from Parser)
      map[string]interface{} (build variables from CLI)

EXAMPLE Variables:
{
  "MODE": "prod",
  "VERSION": "1.2.3"
}
```

**Output:**
```
TYPE: []ResolvedCommand

ResolvedCommand Structure:
{
  Type:      CommandType  // FROM, RUN, COPY, WORKDIR, etc.
  Arguments: []string     // Fully resolved arguments
  Metadata:  CommandMetadata {
    SourceLine:   int     // Original line in Docklett file
    LoopIteration: int    // If from loop, which iteration
  }
}

EXAMPLE:
Input Docklett:
  @if MODE == "prod"
  RUN echo "Production"
  @end
  @for pkg in ["curl", "git"]
  RUN apt-get install -y {{pkg}}
  @end

Input Variables:
  MODE = "prod"

Output:
[
  {Type: RUN, Arguments: ["echo", "Production"],        Metadata: {SourceLine: 2}},
  {Type: RUN, Arguments: ["apt-get", "install", "-y", "curl"], Metadata: {SourceLine: 5, LoopIteration: 0}},
  {Type: RUN, Arguments: ["apt-get", "install", "-y", "git"],  Metadata: {SourceLine: 5, LoopIteration: 1}}
]
```

**Variable Scoping Rules:**

```
GLOBAL SCOPE:
- Variables passed via CLI (--var MODE=prod)
- Available everywhere in Docklett file

LOOP SCOPE:
- Loop variable (e.g., 'pkg' in @for pkg in [...])
- Only available within loop body
- Shadows global variables with same name
- Destroyed after loop ends

Variable Resolution Order:
1. Loop scope (if inside loop)
2. Global scope
3. Error: undefined variable
```

**Type System (V1):**

```
string:  "value"
number:  123, 45.67
boolean: true, false
array:   ["a", "b", "c"]

Type Coercion Rules:
- String comparison: lexicographic
- Number comparison: numeric
- Boolean comparison: true > false
- Array iteration: must be array type
- No implicit type conversion (strict typing)
```

**Evaluation Algorithm:**

```
ALGORITHM: EvaluateAST(node, scope)

CASE node is ProgramNode:
  FOR each statement in node.Statements:
    commands += EvaluateAST(statement, scope)
  RETURN commands

CASE node is DockerCommandNode:
  resolvedArgs = ResolveTemplates(node.Arguments, scope)
  RETURN [{Type: node.Command, Arguments: resolvedArgs}]

CASE node is IfStatementNode:
  conditionValue = EvaluateExpression(node.Condition, scope)
  IF conditionValue is true:
    RETURN EvaluateAST(node.ThenBranch, scope)
  ELSE IF node.ElseBranch exists:
    RETURN EvaluateAST(node.ElseBranch, scope)
  ELSE:
    RETURN []

CASE node is ForLoopNode:
  iterableValue = EvaluateExpression(node.Iterable, scope)
  IF iterableValue is not array:
    ERROR: "for loop requires array"
  
  commands = []
  FOR each element in iterableValue:
    newScope = scope.Clone()
    newScope[node.Variable] = element
    commands += EvaluateAST(node.Body, newScope)
  RETURN commands

CASE node is BinaryExpressionNode:
  left = EvaluateExpression(node.Left, scope)
  right = EvaluateExpression(node.Right, scope)
  RETURN ApplyOperator(node.Operator, left, right)

CASE node is IdentifierNode:
  IF node.Name in scope:
    RETURN scope[node.Name]
  ELSE:
    ERROR: "undefined variable: " + node.Name

CASE node is LiteralNode:
  RETURN node.Value
```

**Template Resolution:**

```
Input:  "apt-get install -y {{pkg}}"
Scope:  {pkg: "curl"}
Output: "apt-get install -y curl"

Implementation:
1. Scan argument string for {{ and }}
2. Extract variable name between delimiters
3. Look up variable in current scope
4. Replace template with string value
5. Repeat until no templates remain
```

**Semantic Error Examples:**
```
Error: Undefined variable 'PKG'
  --> Docklett:12:20
   |
12 | RUN apt-get install {{PKG}}
   |                       ^^^ variable not defined

Error: Type mismatch in comparison
  --> Docklett:3:8
   |
 3 | @if VERSION > "1.0"
   |     ^^^^^^^^^^^^^^^ cannot compare number to string

Error: For loop requires array
  --> Docklett:7:15
   |
 7 | @for x in "not-an-array"
   |           ^^^^^^^^^^^^^^ expected array, got string
```

---

### 5. LLB Builder

**Purpose**: Transform resolved commands into BuildKit LLB (Low-Level Build) graph definition.

**Responsibilities:**
- Convert Docker commands to LLB operations
- Build directed acyclic graph (DAG) of build steps
- Encode dependencies between operations
- Optimize for parallelization opportunities
- Serialize LLB to protobuf or JSON format
- Maintain compatibility with BuildKit API

**Input:**
```
TYPE: []ResolvedCommand (from Interpreter)
```

**Output:**
```
TYPE: llb.Definition (BuildKit LLB Definition)

LLB Structure (Simplified):
{
  Def: [][]byte          // Serialized operation definitions
  Metadata: map[string]OpMetadata {
    Digest: string       // Operation content hash
    Description: string  // Human-readable description
    SourceInfo: {
      Filename: string
      Data: []byte       // Original Docklett source
      Definition: []Range // Line mappings
    }
  }
}

The LLB graph is a Directed Acyclic Graph (DAG) where:
- Nodes = Operations (exec, source, copy, etc.)
- Edges = Data dependencies (output of one op feeds into input of another)
```

**LLB Operation Types (BuildKit SDK):**

```go
// Image source (FROM command)
llb.Image(ref string) llb.State
  Example: llb.Image("ubuntu:22.04")

// Execute command (RUN command)
llb.State.Run(opts ...RunOption) ExecState
  Example: state.Run(llb.Shlex("apt-get update"))

// File operations (COPY command)
llb.State.File(action *FileAction) llb.State
  Example: state.File(llb.Copy(src, srcPath, destPath))

// Working directory (WORKDIR command)
llb.State.Dir(path string) llb.State
  Example: state.Dir("/app")

// Environment variables (ENV command)
llb.State.AddEnv(key, value string) llb.State
  Example: state.AddEnv("PATH", "/usr/local/bin:$PATH")
```

**Conversion Algorithm:**

```
ALGORITHM: BuildLLB(resolvedCommands)

state = nil  // Current build state

FOR each command in resolvedCommands:
  SWITCH command.Type:
  
  CASE "FROM":
    imageName = command.Arguments[0]
    state = llb.Image(imageName)
    AddMetadata(state, command.Metadata)
  
  CASE "RUN":
    commandLine = JoinArguments(command.Arguments)
    state = state.Run(llb.Shlex(commandLine)).Root()
    AddMetadata(state, command.Metadata)
  
  CASE "COPY":
    source = command.Arguments[0]
    dest = command.Arguments[1]
    contextState = llb.Local("context")  // Build context
    state = state.File(llb.Copy(contextState, source, dest))
    AddMetadata(state, command.Metadata)
  
  CASE "WORKDIR":
    path = command.Arguments[0]
    state = state.Dir(path)
    AddMetadata(state, command.Metadata)
  
  CASE "ENV":
    key = command.Arguments[0]
    value = command.Arguments[1]
    state = state.AddEnv(key, value)
    AddMetadata(state, command.Metadata)
  
  // ... other Dockerfile commands

definition = state.Marshal()
RETURN definition
```

**Example Transformation:**

```
Input ResolvedCommands:
[
  {Type: FROM, Arguments: ["ubuntu:22.04"]},
  {Type: RUN, Arguments: ["apt-get", "update"]},
  {Type: RUN, Arguments: ["apt-get", "install", "-y", "curl"]},
  {Type: COPY, Arguments: [".", "/app"]},
  {Type: WORKDIR, Arguments: ["/app"]}
]

Output LLB Graph (conceptual):

┌─────────────────┐
│ Image Source    │
│ ubuntu:22.04    │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Exec: RUN       │
│ apt-get update  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Exec: RUN       │
│ apt-get install │
└────────┬────────┘
         │
         ▼
┌─────────────────┐      ┌─────────────────┐
│ File: COPY      │◀─────│ Local Context   │
│ . → /app        │      │ (source)         │
└────────┬────────┘      └──────────────────┘
         │
         ▼
┌─────────────────┐
│ Dir: WORKDIR    │
│ /app            │
└─────────────────┘
```

**Metadata Encoding:**

```go
// Source mapping for error reporting
sourceInfo := pb.SourceInfo{
  Filename: "Docklett",
  Data:     []byte(originalSource),
  Definition: []*pb.Range{
    {
      Start: pb.Position{Line: 5, Character: 1},
      End:   pb.Position{Line: 5, Character: 30},
    },
  },
}

// Attach to operation
metadata := llb.WithDescription(map[string]string{
  "com.docker.docklett.source": "Docklett:5",
  "com.docker.docklett.loop":   "iteration 2",
})
```

**Optimization Opportunities (Future):**

```
1. Parallel RUN commands that don't depend on each other
   Example:
     RUN apt-get install curl    ┐
                                 ├─ Can run in parallel
     RUN wget https://file.tar   ┘

2. Merge sequential RUN commands into single layer
   Before: RUN a && RUN b && RUN c
   After:  RUN a && b && c

3. Cache mount annotations for package managers
   RUN --mount=type=cache,target=/var/cache/apt apt-get install
```

---

### 6. Build Executor

**Purpose**: Send LLB graph to BuildKit daemon and execute the build.

**Responsibilities:**
- Establish gRPC connection to BuildKit daemon
- Upload build context (files to COPY)
- Submit LLB definition via Solve API
- Stream build progress logs to user
- Handle authentication for registries
- Export final image to Docker or registry
- Report build success/failure

**Input:**
```
TYPE: 
  - llb.Definition (LLB graph from Builder)
  - string (build context directory)
  - BuildOptions (output format, image name, tags, etc.)

BuildOptions:
{
  ImageName:     string
  Tags:          []string
  PushRegistry:  bool
  ExportDocker:  bool        // Export to local Docker daemon
  Platform:      string       // linux/amd64, linux/arm64, etc.
  CacheFrom:     []string     // Import cache from image/registry
  CacheTo:       string       // Export cache to registry
  Secrets:       map[string][]byte
  BuildArgs:     map[string]string
}
```

**Output:**
```
TYPE: BuildResult

BuildResult:
{
  ImageDigest:   string      // sha256:abc123...
  ExporterResponse: map[string]string
  BuildLogs:     []string
  Error:         error
}
```

**Build Execution Flow:**

```
┌────────────────┐
│ 1. Connect     │  Dial BuildKit gRPC (unix socket or TCP)
└───────┬────────┘
        │
        ▼
┌────────────────┐
│ 2. Prepare     │  Package build context into tarball
│    Context     │  Compute content hashes
└───────┬────────┘
        │
        ▼
┌────────────────┐
│ 3. Call Solve  │  client.Solve(ctx, definition, options)
└───────┬────────┘
        │
        ▼
┌────────────────┐
│ 4. Stream Logs │  Display progress: [+] Building 2.3s
│                │  Show execution steps
└───────┬────────┘
        │
        ▼
┌────────────────┐
│ 5. Export      │  Export to Docker daemon (docker load)
│    Result      │  Or push to registry
└───────┬────────┘
        │
        ▼
┌────────────────┐
│ 6. Return      │  Return image digest and metadata
│    Digest      │
└────────────────┘
```

**BuildKit Client API (Go SDK):**

```go
import (
  "github.com/moby/buildkit/client"
  "github.com/moby/buildkit/client/llb"
  "github.com/moby/buildkit/session"
  "github.com/moby/buildkit/frontend/gateway/grpcclient"
)

// Connect to BuildKit
c, err := client.New(ctx, "unix:///var/run/buildkit/buildkitd.sock")
if err != nil {
  return err
}

// Prepare build context
localContext := llb.Local("context",
  llb.SessionID(sessionID),
  llb.SharedKeyHint("context"),
)

// Solve (execute build)
solveOpt := client.SolveOpt{
  Exports: []client.ExportEntry{
    {
      Type: client.ExporterDocker,
      Attrs: map[string]string{
        "name": "myapp:latest",
      },
    },
  },
  LocalDirs: map[string]string{
    "context": "/path/to/build/context",
  },
  Session: []session.Attachable{
    // Auth providers, secret providers, etc.
  },
}

ch := make(chan *client.SolveStatus)
go displayProgress(ch)  // Print build logs

result, err := c.Solve(ctx, definition, solveOpt, ch)
```

**Progress Display (Client-side):**

```
Output format (similar to Docker):

[+] Building 12.5s (8/10)
 => [internal] load build definition from Docklett                 0.1s
 => => transferring dockerfile: 234B                               0.0s
 => [internal] load .dockerignore                                  0.1s
 => => transferring context: 2B                                    0.0s
 => [internal] load metadata for docker.io/library/ubuntu:22.04    1.2s
 => [1/5] FROM docker.io/library/ubuntu:22.04                      0.5s
 => [2/5] RUN apt-get update                                       3.2s
 => [3/5] RUN apt-get install -y curl                              5.8s
 => [4/5] COPY . /app                                              0.3s
 => [5/5] WORKDIR /app                                             0.1s
 => exporting to docker image format                               1.2s
 => => exporting layers                                            0.8s
 => => writing image sha256:abc123...                              0.3s
 => => naming to docker.io/library/myapp:latest                    0.1s
```

**Error Handling:**

```
Connection Errors:
- BuildKit daemon not running
  → Error: failed to connect: connection refused
  → Suggestion: Start BuildKit with `buildkitd` or Docker Desktop

Authentication Errors:
- Registry login required
  → Error: failed to push: authentication required
  → Suggestion: Run `docker login` or provide credentials

Build Errors:
- RUN command failed (exit code 1)
  → Error: executor failed running [/bin/sh -c apt-get install nonexistent]
  → Show full command output from BuildKit
  → Map back to original Docklett line number
```

---

## End-to-End Data Flow

### Complete Compilation Pipeline

```
INPUT: Docklett File
───────────────────────────────────────────────────────────────────
1. CLI Entry Point
   │
   ├─→ Read file: "Docklett"
   ├─→ Parse flags: --var MODE=prod
   └─→ Load build context: "."
   
───────────────────────────────────────────────────────────────────
2. Scanner (Lexical Analysis)
   │
   Input:  FROM ubuntu:22.04\n@if MODE == "prod"\nRUN echo "prod"\n@end
   │
   Process:
   ├─→ Character stream → Token recognition
   ├─→ Track line/column positions
   └─→ Handle string literals, identifiers, operators
   │
   Output: [
     {Type: DOCKER_COMMAND, Lexeme: "FROM", Line: 1},
     {Type: DOCKER_ARGS, Lexeme: "ubuntu:22.04", Line: 1},
     {Type: DIRECTIVE_IF, Lexeme: "@if", Line: 2},
     {Type: IDENTIFIER, Lexeme: "MODE", Line: 2},
     {Type: EQUAL_EQUAL, Lexeme: "==", Line: 2},
     {Type: STRING, Literal: "prod", Line: 2},
     ...
   ]

───────────────────────────────────────────────────────────────────
3. Parser (Syntax Analysis)
   │
   Input:  Token stream
   │
   Process:
   ├─→ Apply grammar rules (recursive descent)
   ├─→ Build hierarchical structure
   └─→ Validate syntax
   │
   Output: AST
   ProgramNode {
     Statements: [
       DockerCommandNode{Command: "FROM", Args: ["ubuntu:22.04"]},
       IfStatementNode{
         Condition: BinaryExpr{
           Left: Identifier("MODE"),
           Op: EQUAL_EQUAL,
           Right: Literal("prod")
         },
         ThenBranch: [
           DockerCommandNode{Command: "RUN", Args: ["echo", "prod"]}
         ]
       }
     ]
   }

───────────────────────────────────────────────────────────────────
4. Interpreter (Semantic Analysis + Evaluation)
   │
   Input:  AST + Variables{MODE: "prod"}
   │
   Process:
   ├─→ Traverse tree
   ├─→ Evaluate MODE == "prod" → true
   ├─→ Expand if branch (discard else)
   ├─→ Unroll for loops
   └─→ Resolve {{template}} variables
   │
   Output: Resolved Commands
   [
     {Type: FROM, Args: ["ubuntu:22.04"]},
     {Type: RUN, Args: ["echo", "prod"]}
   ]

───────────────────────────────────────────────────────────────────
5. LLB Builder (Code Generation)
   │
   Input:  Resolved Commands
   │
   Process:
   ├─→ state = llb.Image("ubuntu:22.04")
   ├─→ state = state.Run(llb.Shlex("echo prod"))
   └─→ definition = state.Marshal()
   │
   Output: LLB Definition (protobuf)
   {
     Def: [<binary operation data>],
     Metadata: {
       "sha256:op1": {Description: "FROM ubuntu:22.04"},
       "sha256:op2": {Description: "RUN echo prod"}
     }
   }

───────────────────────────────────────────────────────────────────
6. Build Executor (Runtime)
   │
   Input:  LLB Definition + Build Context
   │
   Process:
   ├─→ Connect to BuildKit daemon
   ├─→ Package context directory → tarball
   ├─→ client.Solve(ctx, definition, opts)
   ├─→ Stream progress logs
   └─→ Export to Docker daemon
   │
   Output: Container Image
   {
     ImageID: "sha256:abc123...",
     Tags: ["myapp:latest"]
   }

───────────────────────────────────────────────────────────────────
FINAL OUTPUT: Docker image ready to run
$ docker run myapp:latest
```

---

## Missing Pieces & Implementation Details

### 1. Variable Declaration Syntax

**Decision Required:** How are variables declared in Docklett?

**Option A: Implicit (from CLI only)**
```bash
docklett build --var MODE=prod --var VERSION=1.2.3
```
Variables are only defined externally, not in Docklett file.

**Option B: Explicit Declaration (@var directive)**
```dockerfile
@var MODE = "dev"  # Default value
@var VERSION = "0.0.0"

FROM ubuntu:22.04
```
Variables can be declared in file with defaults, overridable by CLI.

**Recommendation:** Option B for better self-documentation.

**Scoping:** All declared variables are global. Loop variables are local to loop body.

---

### 2. String Interpolation in Arguments

**Question:** Should template variables be allowed in all Docker command arguments?

**Example:**
```dockerfile
@var APP_DIR = "/opt/myapp"

WORKDIR {{APP_DIR}}
COPY . {{APP_DIR}}
RUN echo "Installing to {{APP_DIR}}"
```

**Implementation:**
- Scanner recognizes `{{` and `}}` as special tokens
- Parser includes template expressions in argument nodes
- Interpreter resolves templates during command expansion

**Edge Cases:**
```dockerfile
# Multiple variables in one argument
RUN echo "{{USER}}@{{HOST}}"

# Variable in middle of string
COPY app {{VERSION}}-config.yml /etc/
```

---

### 3. Array Functions and Iteration

**V1 Scope:** Only hardcoded arrays
```dockerfile
@for pkg in ["curl", "git", "vim"]
RUN apt-get install {{pkg}}
@end
```

**Future Enhancement:** Array variables
```dockerfile
@var PACKAGES = ["curl", "git", "vim"]

@for pkg in PACKAGES
RUN apt-get install {{pkg}}
@end
```

**Future Enhancement:** Range iteration
```dockerfile
@for i in range(1, 5)
RUN echo "Step {{i}}"
@end
```

---

### 4. Error Recovery Strategy

**Compilation Errors:**

```
┌──────────────────────────────────────────┐
│ Error Severity Levels                    │
├──────────────────────────────────────────┤
│ FATAL:   Stops compilation immediately   │
│          (file not found, syntax error)  │
│                                          │
│ ERROR:   Reportable but recoverable      │
│          (undefined variable - continue  │
│           for multiple error reporting)  │
│                                          │
│ WARNING: Non-blocking issues             │
│          (unused variables)              │
└──────────────────────────────────────────┘
```

**Error Reporting Format:**
```
Error: Undefined variable 'MODE'
  --> Docklett:12:8
   |
12 | @if MODE == "prod"
   |     ^^^^ variable not found
   |
   = help: Define variable with `@var MODE = "value"` or pass --var MODE=value

Error: Expected @end directive
  --> Docklett:18:1
   |
15 | @if CONDITION
16 |   RUN command
17 | @else
18 | @if ANOTHER
   | ^^^ expected @end for @if at line 15
```

**Multiple Error Reporting:**
- Scanner: Continue after lexical errors to find more issues
- Parser: Use panic mode recovery to resynchronize at statement boundaries
- Interpreter: Validate all variable references before evaluation

---

### 5. BuildKit Integration Details

**BuildKit Session Management:**

```go
// Session provides build context and secrets
sess, err := session.NewSession(ctx, "docklett-session", "")
```

**Context Transfer:**
```go
// Efficient streaming of build context
localDirs := map[string]string{
  "context": "/path/to/context",
  "dockerfile": "<inline>",  // LLB definition
}
```

**Cache Configuration:**
```go
// Import cache from previous builds
cacheImports := []client.CacheOptionsEntry{
  {
    Type: "registry",
    Attrs: map[string]string{
      "ref": "myregistry.com/myapp:cache",
    },
  },
}
```

**Progress Display:**
```go
// Vertex = build step
// Progress = execution status
type StatusUpdate struct {
  Vertex:   string  // "FROM ubuntu:22.04"
  Name:     string  // Human-readable
  Total:    int64   // Bytes
  Current:  int64   // Bytes completed
  Timestamp: time.Time
}
```

---

### 6. Testing Strategy

**Unit Tests:**

```
Scanner Tests:
├─ Test token recognition for each type
├─ Test string literal edge cases (escapes, quotes)
├─ Test position tracking accuracy
└─ Test error reporting for invalid characters

Parser Tests:
├─ Test valid programs produce correct AST
├─ Test syntax error detection and reporting
├─ Test operator precedence and associativity
└─ Test nested structures (if inside for, etc.)

Interpreter Tests:
├─ Test variable resolution and scoping
├─ Test conditional evaluation (true/false branches)
├─ Test loop expansion (0, 1, N iterations)
├─ Test template substitution
└─ Test type checking and coercion

LLB Builder Tests:
├─ Test each Docker command maps to correct LLB op
├─ Test state threading (each op feeds into next)
├─ Test metadata preservation
└─ Test serialization correctness
```

**Integration Tests:**

```
End-to-End Tests:
├─ Compile simple Docklett → verify LLB output
├─ Build with BuildKit → verify image contents
├─ Test with real Docker commands (RUN, COPY, etc.)
└─ Test error propagation from BuildKit

Golden File Tests:
├─ Input: Docklett files
├─ Output: Expected LLB JSON (human-readable)
├─ Compare: diff actual vs expected
└─ Update: when intentionally changing behavior
```

**Test Cases:**

```
test_cases/
├─ 01_basic_dockerfile.docklett
│  └─ expected_llb.json
├─ 02_if_else.docklett
│  └─ expected_llb.json
├─ 03_for_loop.docklett
│  └─ expected_llb.json
├─ 04_nested_control_flow.docklett
│  └─ expected_llb.json
├─ 05_template_vars.docklett
│  └─ expected_llb.json
└─ errors/
   ├─ undefined_variable.docklett
   ├─ missing_end.docklett
   └─ type_mismatch.docklett
```

---

### 7. CLI Implementation Details

**Command Structure:**

```bash
docklett [global-options] <command> [command-options] [arguments]

Global Options:
  --verbose, -v     Enable verbose logging
  --debug           Enable debug output (AST, tokens, LLB)
  --help, -h        Show help

Commands:
  build             Compile and build Docklett file
  compile           Compile to LLB without building
  validate          Syntax check only
  version           Show version information
```

**Build Command:**

```bash
docklett build [options] <context-dir>

Options:
  -f, --file FILE          Docklett file (default: ./Docklett)
  -t, --tag TAG            Image name and tag (can specify multiple)
  --var KEY=VALUE          Set build variable (can specify multiple)
  --target STAGE           Build specific stage (for multi-stage)
  --platform PLATFORM      Target platform (linux/amd64, linux/arm64)
  --push                   Push image to registry after build
  --cache-from IMAGE       Import cache from image
  --cache-to DEST          Export cache to destination
  --progress MODE          Progress output mode (auto, plain, tty)
  --no-cache               Disable cache
  --builder NAME           BuildKit builder instance to use

Examples:
  docklett build -f Docklett -t myapp:latest .
  docklett build --var MODE=prod --var VERSION=1.0.0 .
  docklett build --platform linux/amd64,linux/arm64 --push .
```

**Configuration File Support (Future):**

```yaml
# docklett.yml
version: 1
build:
  file: Docklett
  context: .
  variables:
    MODE: dev
    VERSION: 0.0.0-dev
  tags:
    - myapp:dev
  cache:
    from:
      - type: registry
        ref: myregistry.com/myapp:cache
```

---

## Technology Stack

**Language:** Go 1.21+

**Core Dependencies:**
- `github.com/moby/buildkit/client` - BuildKit client SDK
- `github.com/moby/buildkit/frontend/gateway` - LLB construction
- `github.com/moby/buildkit/solver/pb` - LLB protobuf definitions
- `github.com/urfave/cli/v2` - CLI framework
- `github.com/stretchr/testify` - Testing utilities

**Project Structure:**

```
docklett/
├─ cmd/
│  └─ docklett/
│     └─ main.go                # CLI entry point
├─ pkg/
│  ├─ scanner/
│  │  ├─ scanner.go             # Lexical analysis
│  │  └─ token.go               # Token definitions
│  ├─ parser/
│  │  ├─ parser.go              # Syntax analysis
│  │  └─ ast.go                 # AST node definitions
│  ├─ interpreter/
│  │  ├─ interpreter.go         # Semantic analysis
│  │  └─ evaluator.go           # Expression evaluation
│  ├─ builder/
│  │  └─ llb_builder.go         # LLB construction
│  ├─ executor/
│  │  └─ buildkit_client.go     # BuildKit integration
│  └─ errors/
│     └─ errors.go              # Error types and formatting
├─ test/
│  ├─ fixtures/                 # Test Docklett files
│  └─ golden/                   # Expected outputs
├─ go.mod
├─ go.sum
└─ README.md
```

---

## Future Enhancements (Post-V1)

### Advanced Features

**1. Functions**
```dockerfile
@func install_packages(pkgs)
  @for pkg in pkgs
    RUN apt-get install -y {{pkg}}
  @end
@end

@call install_packages(["curl", "git", "vim"])
```

**2. Multi-Stage Builds**
```dockerfile
@var STAGE = "builder"

FROM golang:1.21 AS builder
WORKDIR /src
COPY . .
RUN go build -o app

@if STAGE == "final"
FROM alpine:latest
COPY --from=builder /src/app /app
@end
```

**3. Includes**
```dockerfile
@include "base.docklett"
@include "packages-{{ENVIRONMENT}}.docklett"
```

**4. String Functions**
```dockerfile
@var VERSION = "1.2.3"
@var TAG = "${VERSION}-alpine"  # String interpolation

RUN echo "{{upper(NAME)}}"      # String manipulation
```

**5. Conditional Expressions (Ternary)**
```dockerfile
@var IMAGE = MODE == "prod" ? "alpine:latest" : "alpine:edge"
FROM {{IMAGE}}
```

---

## Open Questions

1. **Variable Mutability:** Should variables be reassignable?
   ```dockerfile
   @var COUNT = 5
   @var COUNT = 10  # Error or allowed?
   ```

2. **Type System:** Should we enforce strict typing or allow dynamic types?

3. **Build-Time vs Runtime:** Are all evaluations compile-time only, or can variables be runtime (ARG/ENV)?

4. **Compatibility:** Should Docklett files be backward-compatible with standard Dockerfiles (ignore @ directives)?

5. **Debugging:** How to debug generated LLB? Should we output intermediate Dockerfile representation?

6. **Performance:** Should we cache AST/LLB for incremental compilation?

---

## Success Metrics (V1)

- ✅ Successfully compile Docklett with variables, if/else, for loops
- ✅ Generate valid LLB accepted by BuildKit
- ✅ Build and run resulting container images
- ✅ Error messages show correct line numbers from original Docklett file
- ✅ Performance comparable to native Dockerfile builds (no significant overhead)
- ✅ Pass all golden file tests (100% of test suite)

---

## References

- [Crafting Interpreters](https://craftinginterpreters.com/) - Compiler design principles
- [BuildKit Documentation](https://docs.docker.com/build/buildkit/) - Build architecture
- [BuildKit LLB](https://github.com/moby/buildkit#exploring-llb) - Low-Level Build spec
- [Dockerfile Reference](https://docs.docker.com/engine/reference/builder/) - Docker commands
