package ast

import (
	"fmt"
	"strings"

	"github.com/lowercasename/golox/token"
)

/* Expressions */

type Expr interface {
	String() string
}

type Assign struct {
	Expr
	Name  token.Token
	Value Expr
}

type Binary struct {
	Expr
	Left     Expr
	Operator token.Token
	Right    Expr
}

// Call expression, for calling a function
type Call struct {
	Expr
	Callee    Expr        // The function being called
	Paren     token.Token // The opening parenthesis
	Arguments []Expr      // The arguments to the function
}

type Grouping struct {
	Expr
	Expression Expr
}

type Literal struct {
	Expr
	Value any
}

type Logical struct {
	Expr
	Left     Expr
	Operator token.Token
	Right    Expr
}

type Unary struct {
	Expr
	Operator token.Token
	Right    Expr
}

// Variable expression, for accessing a variable
type Variable struct {
	Expr
	Name token.Token
}

/* Statements */

type Stmt interface {
	String() string
}

type Expression struct {
	Expr
	Expression Expr
}

// Function statement, for declaring a function
type Function struct {
	Stmt
	Name       token.Token
	Parameters []token.Token
	Body       []Stmt
}

type Block struct {
	Expr
	Statements []Stmt
}

type If struct {
	Expr
	Condition Expr
	Then      Stmt
	Else      Stmt
}

type Print struct {
	Expr
	Expression Expr
}

type Return struct {
	Stmt
	Keyword token.Token
	Value   Expr
}

// Variable statement, for declaring a variable
type Var struct {
	Stmt
	Name        token.Token
	Initializer Expr
}

type While struct {
	Stmt
	Condition Expr
	Body      Stmt
}

/* Printers */

func (a *Assign) String() string {
	return fmt.Sprintf("%s = %s", a.Name.Lexeme, a.Value.String())
}

func (b *Binary) String() string {
	return fmt.Sprintf("(%v %v %v)", b.Operator.Lexeme, b.Left.String(), b.Right.String())
}

func (g *Grouping) String() string {
	return fmt.Sprintf("(group %v)", g.Expression.String())
}

func (l *Literal) String() string {
	return fmt.Sprintf("'%v'", l.Value)
}

func (u *Unary) String() string {
	return fmt.Sprintf("(%v %v)", u.Operator.Lexeme, u.Right.String())
}

func (v *Variable) String() string {
	return fmt.Sprintf("%v", v.Name.Lexeme)
}

func (e *Expression) String() string {
	return fmt.Sprintf("(expression %v)", e.Expression.String())
}

func (p *Print) String() string {
	return fmt.Sprintf("(print %v)", p.Expression.String())
}

func (v *Var) String() string {
	if v.Initializer != nil {
		return fmt.Sprintf("(var %v = %v)", v.Name.Lexeme, v.Initializer.String())
	} else {
		return fmt.Sprintf("(var %v)", v.Name.Lexeme)
	}
}

func (w *While) String() string {
	return fmt.Sprintf("(while %v %v)", w.Condition.String(), w.Body.String())
}

func (i *If) String() string {
	if i.Else != nil {
		return fmt.Sprintf("(if %v %v %v)", i.Condition.String(), i.Then.String(), i.Else.String())
	} else {
		return fmt.Sprintf("(if %v %v)", i.Condition.String(), i.Then.String())
	}
}

func (b *Block) String() string {
	return fmt.Sprintf("(block %v)", b.Statements)
}

func (l *Logical) String() string {
	return fmt.Sprintf("(%v %v %v)", strings.ToUpper(l.Operator.Lexeme), l.Left.String(), l.Right.String())
}

func (f *Function) String() string {
	return fmt.Sprintf("(fun %v %v %v)", f.Name.Lexeme, f.Parameters, f.Body)
}

func (c *Call) String() string {
	return fmt.Sprintf("(call %v %v)", c.Callee.String(), c.Arguments)
}
