package cli

import (
	"flag"
	"fmt"
	"os"
)

type CommandLine struct {
	FilePath string
}

func NewCommandLine() *CommandLine {
	return &CommandLine{}
}

func (c *CommandLine) ParseArgs() error {
	flag.StringVar(&c.FilePath, "file", "", "Path to Dockerfile or Docklett file")
	flag.StringVar(&c.FilePath, "F", "", "Path to Dockerfile or Docklett file (shorthand)")
	flag.Parse()

	if c.FilePath == "" {
		if flag.NArg() > 0 {
			c.FilePath = flag.Arg(0)
		} else {
			return fmt.Errorf("file path is required")
		}
	}

	if _, err := os.Stat(c.FilePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", c.FilePath)
	}

	return nil
}
