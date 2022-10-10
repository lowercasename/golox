package ast

import (
	"fmt"

	"github.com/lowercasename/golox/token"
)

type Expr interface {
	String() string
}

type Binary struct {
	Expr
	Left     Expr
	Operator token.Token
	Right    Expr
}

type Grouping struct {
	Expr
	Expression Expr
}

type Literal struct {
	Expr
	Value any
}

type Unary struct {
	Expr
	Operator token.Token
	Right    Expr
}

/* Printers */

func (b *Binary) String() string {
	return fmt.Sprintf("(%v %v %v)", b.Operator.Lexeme, b.Left.String(), b.Right.String())
}

func (g *Grouping) String() string {
	return fmt.Sprintf("(group %v)", g.Expression.String())
}

func (l *Literal) String() string {
	return fmt.Sprintf("%v", l.Value)
}

func (u *Unary) String() string {
	return fmt.Sprintf("(%v %v)", u.Operator.Lexeme, u.Right.String())
}
