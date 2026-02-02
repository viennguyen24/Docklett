package main

import (
	"docklett/cli"
	"docklett/compiler"
	"fmt"
	"os"
)

func main() {
	commandLine := cli.NewCommandLine()
	err := commandLine.ParseArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	comp := compiler.NewCompiler()
	err = comp.Run(commandLine.FilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Compilation failed: %v\n", err)
		os.Exit(1)
	}

	if comp.HasError {
		fmt.Fprintf(os.Stderr, "Compilation completed with errors\n")
		os.Exit(1)
	}
}
