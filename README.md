# Docklett

A Golang compiler to add programming language features to Dockerfile syntax.

## Prerequisites

- Go 1.25.5 or higher
- Git (for version control)

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
Check `design/DESIGN.md` for architecture details.

## License

See [LICENSE](LICENSE) file.
