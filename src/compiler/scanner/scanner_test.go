package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"docklett/compiler/token"
)

func tokenTypeName(t token.TokenType) string {
	switch t {
	case token.IDENTIFIER:
		return "IDENTIFIER"
	case token.STRING:
		return "STRING"
	case token.NUMBER:
		return "NUMBER"
	case token.BOOL:
		return "BOOL"
	case token.EQUAL:
		return "EQUAL"
	case token.ASSIGN:
		return "ASSIGN"
	case token.UNEQUAL:
		return "UNEQUAL"
	case token.ADD:
		return "ADD"
	case token.ADD_ASSIGN:
		return "ADD_ASSIGN"
	case token.SUBTRACT:
		return "SUBTRACT"
	case token.SUB_ASSIGN:
		return "SUB_ASSIGN"
	case token.MULTI:
		return "MULTI"
	case token.MULTI_ASSIGN:
		return "MULTI_ASSIGN"
	case token.DIVIDE:
		return "DIVIDE"
	case token.DIV_ASSIGN:
		return "DIV_ASSIGN"
	case token.NEGATE:
		return "NEGATE"
	case token.AND:
		return "AND"
	case token.OR:
		return "OR"
	case token.GREATER:
		return "GREATER"
	case token.LESS:
		return "LESS"
	case token.GTE:
		return "GTE"
	case token.LTE:
		return "LTE"
	case token.LPAREN:
		return "LPAREN"
	case token.RPAREN:
		return "RPAREN"
	case token.LBRACE:
		return "LBRACE"
	case token.RBRACE:
		return "RBRACE"
	case token.LBRACKET:
		return "LBRACKET"
	case token.RBRACKET:
		return "RBRACKET"
	case token.COLON:
		return "COLON"
	case token.COMMA:
		return "COMMA"
	case token.SET:
		return "SET"
	case token.IF:
		return "IF"
	case token.ELIF:
		return "ELIF"
	case token.ELSE:
		return "ELSE"
	case token.FOR:
		return "FOR"
	case token.IN:
		return "IN"
	case token.END:
		return "END"
	case token.TRUE:
		return "TRUE"
	case token.FALSE:
		return "FALSE"
	case token.DLINE:
		return "DLINE"
	case token.NLINE:
		return "NLINE"
	case token.EOF:
		return "EOF"
	case token.ILLEGAL:
		return "ILLEGAL"
	default:
		return fmt.Sprintf("TokenType(%d)", t)
	}
}

func literalString(lit any) string {
	if lit == nil {
		return "nil"
	}
	switch v := lit.(type) {
	case string:
		return fmt.Sprintf("%q", v)
	case int:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case float64:
		return fmt.Sprintf("%g", v)
	case bool:
		if v {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", v)
	}
}

func scanAndPrintTokens(t *testing.T, filename, source string) {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatalf("write temp source: %v", err)
	}

	var s Scanner
	if err := s.ReadSource(path); err != nil {
		t.Fatalf("read source: %v", err)
	}
	if err := s.ScanSource(); err != nil {
		t.Fatalf("scan source: %v", err)
	}

	fmt.Printf("=== %s ===\n", filename)
	for _, tok := range s.Tokens {
		fmt.Printf("%-10s lexeme=%q literal=%s\n", tokenTypeName(tok.Type), tok.Lexeme, literalString(tok.Literal))
	}
}

func TestScan_IfElse(t *testing.T) {
	source := "@IF MODE == \"prod\"\n@ELSE\n@END\n"
	scanAndPrintTokens(t, "if_else.dock", source)
}

func TestScan_ForList(t *testing.T) {
	source := "@FOR pkg IN [\"curl\",\"git\",\"vim\"]\n@END\n"
	scanAndPrintTokens(t, "for_list.dock", source)
}

func TestScan_SetAndBool(t *testing.T) {
	source := "@SET NAME = \"docklett\"\n@IF TRUE\n@END\n"
	scanAndPrintTokens(t, "set_bool.dock", source)
}

func TestScan_DockerfileTemplate(t *testing.T) {
	source := "FROM python:3.12-slim\nWORKDIR /app\nCOPY requirements.txt ./\nRUN pip install --no-cache-dir -r requirements.txt\nCOPY . .\nEXPOSE 8080\nCMD [\"python\", \"app.py\"]\n"
	scanAndPrintTokens(t, "dockerfile_template.dock", source)
}

func TestScan_DockerfileLineContinuation(t *testing.T) {
	source := "RUN echo hello \\\n    && echo world\n"
	scanAndPrintTokens(t, "dockerfile_line_continuation.dock", source)
}

func TestScan_NewlineAfterDocker(t *testing.T) {
	source := "FROM alpine\nRUN echo test\n"
	scanAndPrintTokens(t, "newline_after_docker.dock", source)
}
