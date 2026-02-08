package compiler

import (
	"docklett/compiler/ast"
	"docklett/compiler/parser"
	"docklett/compiler/scanner"
	"docklett/compiler/token"
)

type Compiler struct {
	Scanner         *scanner.Scanner
	Parser          *parser.Parser
	InputFilePath   string
	InputFileName   string
	GeneratedTokens []token.Token
	GeneratedAST    ast.Expression
	HasError        bool
}

func NewCompiler() *Compiler {
	return &Compiler{
		Scanner:  &scanner.Scanner{},
		Parser:   &parser.Parser{},
		HasError: false,
	}
}

// main entry point
func (c *Compiler) Run(inputFilePath string) error {
	c.InputFilePath = inputFilePath

	err := c.Scanner.ReadSource(inputFilePath)
	if err != nil {
		c.HasError = true
		return err
	}

	c.InputFileName = c.Scanner.SourceName

	err = c.Scanner.ScanSource()
	if err != nil {
		c.HasError = true
		return err
	}

	c.GeneratedTokens = c.Scanner.Tokens
	// statements, err := c.Parser.Parse(c.GeneratedTokens)
	// _, err := c.Interpreter.Interpret(statements)

	if err != nil {
		c.HasError = true
		return err
	}

	// c.GeneratedAST = ast

	return nil
}
