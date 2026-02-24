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
	RANGE // built-in function keyword for range(start, end) or range(start, end, step)

	// Docker instruction tokens
	DOCKER_KEYWORD // instruction verb: FROM, RUN, COPY, ENV, etc.
	DOCKER_ARGS    // everything after the keyword on the same logical line

	EOF
	ILLEGAL
)

var DockerTokenKeywords = map[string]TokenType{
	"ADD":         DOCKER_KEYWORD,
	"ARG":         DOCKER_KEYWORD,
	"CMD":         DOCKER_KEYWORD,
	"COPY":        DOCKER_KEYWORD,
	"ENTRYPOINT":  DOCKER_KEYWORD,
	"ENV":         DOCKER_KEYWORD,
	"EXPOSE":      DOCKER_KEYWORD,
	"FROM":        DOCKER_KEYWORD,
	"HEALTHCHECK": DOCKER_KEYWORD,
	"LABEL":       DOCKER_KEYWORD,
	"MAINTAINER":  DOCKER_KEYWORD,
	"ONBUILD":     DOCKER_KEYWORD,
	"RUN":         DOCKER_KEYWORD,
	"SHELL":       DOCKER_KEYWORD,
	"STOPSIGNAL":  DOCKER_KEYWORD,
	"USER":        DOCKER_KEYWORD,
	"VOLUME":      DOCKER_KEYWORD,
	"WORKDIR":     DOCKER_KEYWORD,
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
	"range": RANGE,
}

var TokenTypeNames = map[TokenType]string{
	IDENTIFIER:     "IDENTIFIER",
	STRING:         "STRING",
	NUMBER:         "NUMBER",
	BOOL:           "BOOL",
	EQUAL:          "EQUAL",
	ASSIGN:         "ASSIGN",
	UNEQUAL:        "UNEQUAL",
	ADD:            "ADD",
	ADD_ASSIGN:     "ADD_ASSIGN",
	SUBTRACT:       "SUBTRACT",
	SUB_ASSIGN:     "SUB_ASSIGN",
	MULTI:          "MULTI",
	MULTI_ASSIGN:   "MULTI_ASSIGN",
	DIVIDE:         "DIVIDE",
	DIV_ASSIGN:     "DIV_ASSIGN",
	NEGATE:         "NEGATE",
	AND:            "AND",
	OR:             "OR",
	GREATER:        "GREATER",
	LESS:           "LESS",
	GTE:            "GTE",
	LTE:            "LTE",
	LPAREN:         "LPAREN",
	RPAREN:         "RPAREN",
	LBRACE:         "LBRACE",
	RBRACE:         "RBRACE",
	LBRACKET:       "LBRACKET",
	RBRACKET:       "RBRACKET",
	COLON:          "COLON",
	COMMA:          "COMMA",
	SET:            "SET",
	IF:             "IF",
	ELIF:           "ELIF",
	ELSE:           "ELSE",
	FOR:            "FOR",
	IN:             "IN",
	END:            "END",
	TRUE:           "TRUE",
	FALSE:          "FALSE",
	RANGE:          "RANGE",
	DOCKER_KEYWORD: "DOCKER_KEYWORD",
	DOCKER_ARGS:    "DOCKER_ARGS",
	EOF:            "EOF",
	ILLEGAL:        "ILLEGAL",
	NLINE:          "NEW_LINE",
}

type Token struct {
	Type   TokenType
	Lexeme string
	Position
	Literal any
}
