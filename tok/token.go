package tok

import "fmt"

type Token struct {
    TokenType TokenType
    Lexeme string
    Literal any
    Line int
}

func (token Token) String() string {
    return fmt.Sprintf("%v %v %v", token.TokenType, token.Lexeme, token.Literal)
}
