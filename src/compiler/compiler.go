package compiler

import (
	"docklett/compiler/parser"
	"docklett/compiler/scanner"
	"docklett/compiler/token"
)

type Compiler struct {
	Scanner        *scanner.Scanner
	Parser         *parser.Parser
	InputFilePath  string
	InputFileName  string
	GeneratedTokens []token.Token
	GeneratedAST   parser.Expression
	HasError       bool
}

func NewCompiler() *Compiler {
	return &Compiler{
		Scanner:  &scanner.Scanner{},
		Parser:   &parser.Parser{},
		HasError: false,
	}
}

// main entry point
func (i *Compiler) Run(inputFilePath string) error {
	i.InputFilePath = inputFilePath
	
	err := i.Scanner.ReadSource(inputFilePath)
	if err != nil {
		i.HasError = true
		return err
	}
	
	i.InputFileName = i.Scanner.SourceName
	
	err = i.Scanner.ScanSource()
	if err != nil {
		i.HasError = true
		return err
	}
	
	i.GeneratedTokens = i.Scanner.Tokens
	
	i.Parser.Tokens = i.GeneratedTokens
	
	ast, err := i.Parser.Parse()
	if err != nil {
		i.HasError = true
		return err
	}
	
	i.GeneratedAST = ast
	
	return nil
}


