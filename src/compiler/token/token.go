package token

type Position struct {
	Line int
	File string
	Col  int
}

type TokenType int

const (
	// Identifiers + literals
	_ TokenType = iota
	IDENTIFIER
	STRING
	NUMBER
	BOOL

	// Operators
	EQUAL        //
	ASSIGN       //
	UNEQUAL      //
	ADD          //
	ADD_ASSIGN   //
	SUBTRACT     //
	SUB_ASSIGN   //
	MULTI        //
	MULTI_ASSIGN //
	DIVIDE       //
	DIV_ASSIGN   //
	NEGATE       //
	AND          //
	OR           //
	GREATER      //
	LESS         //
	GTE          //
	LTE          //

	// Delimiters
	LPAREN   //
	RPAREN   //
	LBRACE   //
	RBRACE   //
	LBRACKET //
	RBRACKET //
	COLON    //
	COMMA    //
	NLINE

	// Keywords
	SET
	IF
	ELIF
	ELSE
	FOR
	IN
	END
	TRUE
	FALSE

	// raw line in Dockerfile
	// Docklett should be responsible only for our add-ons
	DLINE

	EOF
	ILLEGAL
)

var DockerTokenKeywords = map[string]TokenType{
	"ADD":         DLINE,
	"ARG":         DLINE,
	"CMD":         DLINE,
	"COPY":        DLINE,
	"ENTRYPOINT":  DLINE,
	"ENV":         DLINE,
	"EXPOSE":      DLINE,
	"FROM":        DLINE,
	"HEALTHCHECK": DLINE,
	"LABEL":       DLINE,
	"MAINTAINER":  DLINE,
	"ONBUILD":     DLINE,
	"RUN":         DLINE,
	"SHELL":       DLINE,
	"STOPSIGNAL":  DLINE,
	"USER":        DLINE,
	"VOLUME":      DLINE,
	"WORKDIR":     DLINE,
}

var DocklettTokenKeywords = map[string]TokenType{
	"SET":   SET,
	"IF":    IF,
	"ELIF":  ELIF,
	"ELSE":  ELSE,
	"FOR":   FOR,
	"IN":    IN,
	"END":   END,
	"TRUE":  TRUE,
	"FALSE": FALSE,
}

var TokenTypeNames = map[TokenType]string{
	IDENTIFIER:   "IDENTIFIER",
	STRING:       "STRING",
	NUMBER:       "NUMBER",
	BOOL:         "BOOL",
	EQUAL:        "EQUAL",
	ASSIGN:       "ASSIGN",
	UNEQUAL:      "UNEQUAL",
	ADD:          "ADD",
	ADD_ASSIGN:   "ADD_ASSIGN",
	SUBTRACT:     "SUBTRACT",
	SUB_ASSIGN:   "SUB_ASSIGN",
	MULTI:        "MULTI",
	MULTI_ASSIGN: "MULTI_ASSIGN",
	DIVIDE:       "DIVIDE",
	DIV_ASSIGN:   "DIV_ASSIGN",
	NEGATE:       "NEGATE",
	AND:          "AND",
	OR:           "OR",
	GREATER:      "GREATER",
	LESS:         "LESS",
	GTE:          "GTE",
	LTE:          "LTE",
	LPAREN:       "LPAREN",
	RPAREN:       "RPAREN",
	LBRACE:       "LBRACE",
	RBRACE:       "RBRACE",
	LBRACKET:     "LBRACKET",
	RBRACKET:     "RBRACKET",
	COLON:        "COLON",
	COMMA:        "COMMA",
	SET:          "SET",
	IF:           "IF",
	ELIF:         "ELIF",
	ELSE:         "ELSE",
	FOR:          "FOR",
	IN:           "IN",
	END:          "END",
	TRUE:         "TRUE",
	FALSE:        "FALSE",
	DLINE:        "DLINE",
	EOF:          "EOF",
	ILLEGAL:      "ILLEGAL",
	NLINE:        "NEW_LINE",
}

type Token struct {
	Type   TokenType
	Lexeme string
	Position
	Literal any
}
