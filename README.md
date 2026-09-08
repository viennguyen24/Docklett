# Docklett

A Golang compiler to add programming language features to Dockerfile syntax.

## Prerequisites

- Go 1.25.5 or higher
- Git (for version control)

## Getting Started

### Step 1: Clone the Repository
```bash
git clone <repository-url>
cd Docklett
```

### Step 2: Verify Go Installation
```bash
go version  # Should show Go 1.25.5 or higher
```

### Step 3: Install Development Tools (Skip if you don't need to work on source code)

**On Unix/Linux/macOS/Git Bash (Windows):**
```bash
chmod +x scripts/setup-hooks.sh
./scripts/setup-hooks.sh
```

**On Windows PowerShell:**
```powershell
.\scripts\setup-hooks.ps1
```

This installs Git hooks that automatically run `go fmt` and `go vet` before each commit.

### Step 4: Verify Setup
```bash
cd src
go build -o ../docklett.exe main.go
```

If successful, you're ready to develop!

## Project Structure

```
Docklett/
├── src/
│   ├── main.go           # Entry point
│   ├── go.mod            # Go module definition
│   ├── cli/              # Command-line interface
│   |── compiler/         # Compiler components
│         └── ...
├── design/               # Design documentation
└── README.md
```

## Building the Project

### Build executable
```bash
cd src
go build -o ../docklett.exe main.go
```

This creates `docklett.exe` in the project root.

## Running the Project

### Option 1: Run directly with Go
```bash
cd src
go run main.go -file <path-to-docklett-file>
```

### Option 2: Run compiled executable
```bash
./docklett.exe -file <path-to-docklett-file>
```

### Command-line flags
- `-file <path>` : Path to Dockerfile or Docklett file
- `-F <path>` : Shorthand for `-file`
- `--help` : Display usage information

## Example Usage

```bash
# Using go run
cd src
go run main.go -file ../example.docklett

# Using compiled binary
./docklett.exe -file example.docklett

# Using shorthand flag
./docklett.exe -F example.docklett
```

## Development

### Running Tests

Compiler tests live under `src/tests/` (module root is `src/`, where `go.mod` is). Shared helpers are in `tests/testutil`; scanner behavioral tests are in `tests/scanner` (see that folder’s `README.md`).

```bash
# From repo root, enter the Go module
cd src

# Run the full compiler test tree
go test ./tests/...

# Scanner suite only
go test ./tests/scanner/

# Shared helpers package
go test ./tests/testutil/

# Verbose
go test -v ./tests/scanner/
```

Run tests from `src/` so package paths resolve against `go.mod`. Do not use the old `compiler/scanner` package path for the suite.
### Installing Git Hooks

This project uses pre-commit hooks to automatically run `go fmt` and `go vet` before each commit.

**Installation (Unix/Linux/macOS/Git Bash on Windows):**
```bash
./scripts/setup-hooks.sh
```

**Installation (Windows PowerShell):**
```powershell
.\scripts\setup-hooks.ps1
```

**What the hook does:**
- Runs `gofmt -w` on each staged `.go` file and re-stages changes (works across packages)
- Runs `go vet ./...` from `src/` where `go.mod` lives
- Blocks commit if vet finds issues

**Bypassing the hook (not recommended):**
```bash
git commit --no-verify
```

### Continuous Integration

Pushes and pull requests to `main` run `.github/workflows/ci.yml`: `gofmt` check, `go vet ./...`, and `go test ./tests/...` from `src/`. Same checks as the local hook (plus tests); CI cannot be skipped with `--no-verify`.

### Architecture

Check `design/DESIGN.md` for architecture details.

## License

See [LICENSE](LICENSE) file.
