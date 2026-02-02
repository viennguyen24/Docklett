package parser

import (
	"docklett/compiler/token"
	"fmt"
	"reflect"
	"strings"
)

var tokenTypeNames = map[token.TokenType]string{
	token.IDENTIFIER: "IDENTIFIER",
	token.STRING:     "STRING",
	token.NUMBER:     "NUMBER",
	token.BOOL:       "BOOL",
	token.EQUAL:      "EQUAL",
	token.ASSIGN:     "ASSIGN",
	token.UNEQUAL:    "UNEQUAL",
	token.ADD:        "ADD",
	token.ADD_ASSIGN: "ADD_ASSIGN",
	token.SUBTRACT:   "SUBTRACT",
	token.SUB_ASSIGN: "SUB_ASSIGN",
	token.MULTI:      "MULTI",
	token.MULTI_ASSIGN: "MULTI_ASSIGN",
	token.DIVIDE:     "DIVIDE",
	token.DIV_ASSIGN: "DIV_ASSIGN",
	token.NEGATE:     "NEGATE",
	token.AND:        "AND",
	token.OR:         "OR",
	token.GREATER:    "GREATER",
	token.LESS:       "LESS",
	token.GTE:        "GTE",
	token.LTE:        "LTE",
	token.LPAREN:     "LPAREN",
	token.RPAREN:     "RPAREN",
	token.LBRACE:     "LBRACE",
	token.RBRACE:     "RBRACE",
	token.LBRACKET:   "LBRACKET",
	token.RBRACKET:   "RBRACKET",
	token.COLON:      "COLON",
	token.COMMA:      "COMMA",
	token.SET:        "SET",
	token.IF:         "IF",
	token.ELIF:       "ELIF",
	token.ELSE:       "ELSE",
	token.FOR:        "FOR",
	token.IN:         "IN",
	token.END:        "END",
	token.TRUE:       "TRUE",
	token.FALSE:      "FALSE",
	token.DLINE:      "DLINE",
	token.EOF:        "EOF",
	token.ILLEGAL:    "ILLEGAL",
}

type TreePrinter struct {
	indentLevel int
	indentChar  string
	isLastChild []bool
}

func NewTreePrinter() *TreePrinter {
	return &TreePrinter{
		indentLevel: 0,
		indentChar:  "  ",
		isLastChild: []bool{},
	}
}

func (tp *TreePrinter) getIndent(isLast bool, isField bool) string {
	var result strings.Builder
	
	for i := 0; i < len(tp.isLastChild); i++ {
		if tp.isLastChild[i] {
			result.WriteString("  ")
		} else {
			result.WriteString("│ ")
		}
	}
	
	if isField {
		if isLast {
			result.WriteString("└─")
		} else {
			result.WriteString("├─")
		}
	}
	
	return result.String()
}


func (tp *TreePrinter) PrintTree(expr Expression) string {
	if expr == nil {
		return "<nil>"
	}

	exprValue := reflect.ValueOf(expr)
	exprType  := reflect.TypeOf(expr)
	
	typeName := exprType.Name()
	result   := typeName + "\n"
	
	numFields := exprValue.NumField()
	
	for i := 0; i < numFields; i++ {
		field := exprValue.Field(i)
		fieldType := exprType.Field(i)
		fieldName := fieldType.Name
		isLast := (i == numFields-1)
		
		prefix := tp.getIndent(isLast, true)
		result += prefix + fieldName + ": "
		
		if field.Type().String() == "token.Token" {
			tok := field.Interface().(token.Token)
			result += tp.formatToken(tok) + "\n"
		} else if field.Type().Kind() == reflect.Interface && field.CanInterface() {
			if exprField, ok := field.Interface().(Expression); ok {
				tp.isLastChild = append(tp.isLastChild, isLast)
				nestedTree := tp.PrintTree(exprField)
				result += nestedTree
				tp.isLastChild = tp.isLastChild[:len(tp.isLastChild)-1]
			} else {
				result += fmt.Sprintf("%v\n", field.Interface())
			}
		} else {
			result += fmt.Sprintf("%v\n", field.Interface())
		}
	}
	
	return result
}

func (tp *TreePrinter) formatToken(tok token.Token) string {
	typeName := tokenTypeNames[tok.Type]
	if typeName == "" {
		typeName = fmt.Sprintf("UNKNOWN(%d)", tok.Type)
	}
	
	base := fmt.Sprintf("%s [%s] @Line:%d,Col:%d", 
		typeName, tok.Lexeme, tok.Position.Line, tok.Position.Col)
	
	if tok.Literal != nil {
		base += fmt.Sprintf(" (Literal: %v)", tok.Literal)
	}
	
	return base
}

func PrintAST(expr Expression) {
	printer := NewTreePrinter()
	fmt.Println(printer.PrintTree(expr))
}

func DemoPrinter() {
	// (5 + 3) * 2
	expr := Binary{
		Left: Grouping{
			Expression: Binary{
				Left: LiteralExpression{Value: 5},
				Operator: token.Token{
					Type:   token.ADD,
					Lexeme: "+",
					Position: token.Position{Line: 1, Col: 3},
				},
				Right: LiteralExpression{Value: 3},
			},
		},
		Operator: token.Token{
			Type:   token.MULTI,
			Lexeme: "*",
			Position: token.Position{Line: 1, Col: 8},
		},
		Right: LiteralExpression{Value: 2},
	}

	PrintAST(expr)
}