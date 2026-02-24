/*
COMPILE-TIME ERRORS occur during source code analysis (scanning and parsing).
These errors are detected before any code execution and indicate syntax or lexical problems.

We will pivot the entire compiler to use Compile Time error. Our Compiler is only responsible for building a correct LLB state,
and the Docker engine will be handling execution. As such all errors in compiler will be compile time and error in Docker engine build will be run-time

Error Types:
  - ScanError: Lexical analysis errors (unknown characters, malformed tokens)
  - ParseError: Syntax analysis errors (unexpected tokens, grammar violations)

EXAMPLES:

	ScanError:  Unknown character '@#' at line 5
	ParseError: Expected ')' after expression at line 10
*/
package error

import (
	"docklett/compiler/token"
	"fmt"
)

type CompileError interface {
	error
	GetLine() int
	GetLocation() string
}

// ScanError represents lexical analysis errors
type ScanError struct {
	Line    int
	Column  int
	File    string
	Message string
}

func (e *ScanError) Error() string {
	return fmt.Sprintf("Compile Error: [line %d] %s", e.Line, e.Message)
}

func (e *ScanError) GetLine() int {
	return e.Line
}

func (e *ScanError) GetLocation() string {
	if e.File != "" {
		return fmt.Sprintf("file %s, line %d, column %d", e.File, e.Line, e.Column)
	}
	return fmt.Sprintf("line %d, column %d", e.Line, e.Column)
}

func NewScanError(line, column int, file, message string) *ScanError {
	return &ScanError{
		Line:    line,
		Column:  column,
		File:    file,
		Message: message,
	}
}

// ParseError represents syntax analysis errors
type ParseError struct {
	Token   token.Token
	Message string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("Compile Error: [line %d] %s", e.Token.Position.Line, e.Message)
}

func (e *ParseError) GetLine() int {
	return e.Token.Position.Line
}

func (e *ParseError) GetLocation() string {
	if e.Token.Type == token.EOF {
		return "at end"
	}
	return fmt.Sprintf("at '%s'", e.Token.Lexeme)
}

func NewParseError(tok token.Token, message string) *ParseError {
	return &ParseError{
		Token:   tok,
		Message: message,
	}
}

// TranslatorError represents AST-to-LLB translation errors
type TranslatorError struct {
	Line    int
	Message string
}

func (e *TranslatorError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("Compile Error: [line %d] %s", e.Line, e.Message)
	}
	return fmt.Sprintf("Compile Error: %s", e.Message)
}

func (e *TranslatorError) GetLine() int {
	return e.Line
}

func (e *TranslatorError) GetLocation() string {
	return fmt.Sprintf("line %d", e.Line)
}

func NewTranslatorError(line int, message string) *TranslatorError {
	return &TranslatorError{
		Line:    line,
		Message: message,
	}
}

// PanicTranslatorError panics with a translator compile error
func PanicTranslatorError(line int, message string) {
	panic(NewTranslatorError(line, message))
}
