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

```bash
# From repo root, navigate to src/ where go.mod is located
cd src

# Run all tests in all packages
go test ./...

# Run tests in specific package
go test ./compiler/scanner
go test ./compiler/parser
go test ./compiler/interpreter

# Run tests with verbose output
go test -v ./...
```

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
- Runs `go fmt` on all staged `.go` files and auto-stages changes
- Runs `go vet ./...` to check for common mistakes
- Blocks commit if vet finds issues

**Bypassing the hook (not recommended):**
```bash
git commit --no-verify
```

### Architecture

Check `design/DESIGN.md` for architecture details.

## License

See [LICENSE](LICENSE) file.
