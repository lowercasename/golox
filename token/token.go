package token

import "fmt"

type Type string

const (
	// single-character tokens
	LEFTPAREN  = "("
	RIGHTPAREN = ")"
	LEFTBRACE  = "{"
	RIGHTBRACE = "}"
	COMMA      = ","
	DOT        = "."
	MINUS      = "-"
	PLUS       = "+"
	SEMICOLON  = ";"
	SLASH      = "/"
	STAR       = "*"
	QMARK      = "?"
	COLON      = ":"
	// one or two character tokens
	BANG         = "!"
	BANGEQUAL    = "!="
	EQUAL        = "="
	EQUALEQUAL   = "=="
	GREATER      = ">"
	GREATEREQUAL = ">="
	LESS         = "<"
	LESSEQUAL    = "<="
	// literals
	IDENTIFIER = "IDENTIFIER"
	STRING     = "STRING"
	NUMBER     = "NUMBER"
	// keywords
	AND     = "and"
	CLASS   = "class"
	ELSE    = "else"
	FALSE   = "false"
	FUN     = "fun"
	FOR     = "for"
	IF      = "if"
	NIL     = "nil"
	OR      = "or"
	PRINT   = "print"
	RETURN  = "return"
	SUPER   = "super"
	THIS    = "this"
	TRUE    = "true"
	VAR     = "var"
	WHILE   = "while"
	EOF     = "EOF"
	INVALID = "__INVALID__"
)

type Token struct {
	Type    Type
	Lexeme  string
	Literal any
	Line    int
}

func (token *Token) String() string {
	return fmt.Sprintf("%s %s %v", token.Type, token.Lexeme, token.Literal)
}
