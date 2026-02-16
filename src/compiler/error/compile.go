/*
COMPILE-TIME ERRORS occur during source code analysis (scanning and parsing).
These errors are detected before any code execution and indicate syntax or lexical problems.

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

// CompileError represents errors detected during compilation (scanning or parsing).
// These errors occur before any code execution.
//
// Functionality:
//   - Extends Go's error interface with line number and location tracking
//   - Enables structured error reporting with precise source locations
//
// Used by: Scanner and Parser to report lexical and syntax errors
type CompileError interface {
	error
	GetLine() int        // Returns the line number where the error occurred
	GetLocation() string // Returns a human-readable location description
}

// ScanError represents lexical analysis errors detected during scanning.
// Occurs when the scanner encounters invalid characters or malformed tokens.
//
// Functionality:
//   - Tracks precise location (file, line, column) of lexical errors
//   - Provides detailed error messages for unknown tokens
//
// Examples:
//
//	Unknown character '@#' at line 5, column 10
//	Unterminated string at line 20, column 15
//
// Used in: Scanner when encountering invalid input characters
type ScanError struct {
	Line    int    // Line number where error occurred (1-indexed)
	Column  int    // Column number where error occurred (1-indexed)
	File    string // Source file name (optional)
	Message string // Human-readable error description
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

// NewScanError creates a new lexical analysis error.
//
// Parameters:
//
//	line: Line number where error occurred (1-indexed)
//	column: Column number where error occurred (1-indexed)
//	file: Source file name (empty string if not applicable)
//	message: Human-readable error description
//
// Example:
//
//	err := NewScanError(10, 5, "main.dock", "Unknown character '@'")
func NewScanError(line, column int, file, message string) *ScanError {
	return &ScanError{
		Line:    line,
		Column:  column,
		File:    file,
		Message: message,
	}
}

// ParseError represents syntax analysis errors detected during parsing.
// Occurs when token sequence doesn't match expected grammar rules.
//
// Functionality:
//   - Associated with specific token where parsing failed
//   - Provides context about what was expected vs. found
//
// Examples:
//
//	Compile Error: [line 10] Expected ')' after expression
//	Compile Error: [line 15] Unexpected token '@SET'
//
// Used in: Parser when encountering syntax violations or unexpected tokens
type ParseError struct {
	Token   token.Token // Token where parsing error occurred
	Message string      // Human-readable error description
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

// NewParseError creates a new syntax analysis error.
//
// Parameters:
//
//	tok: Token where the parsing error occurred
//	message: Human-readable error description
//
// Example:
//
//	err := NewParseError(currentToken, "Expected ';' after statement")
//	err := NewParseError(currentToken, "Unable to perform assignment on expression")
func NewParseError(tok token.Token, message string) *ParseError {
	return &ParseError{
		Token:   tok,
		Message: message,
	}
}
