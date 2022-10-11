package parser

import (
	"fmt"

	"github.com/lowercasename/golox/ast"
	"github.com/lowercasename/golox/logger"
	"github.com/lowercasename/golox/token"
)

type Parser struct {
	tokens  []token.Token
	current int
}

func New(tokens []token.Token) Parser {
	return Parser{tokens, 0}
}

// Start parsing
func (parser *Parser) Parse() []ast.Expr {
	var expressions []ast.Expr
	for !parser.isAtEnd() {
		expr, err := parser.expression()
		if err != nil {
			fmt.Println(err)
			parser.synchronize()
		} else {
			expressions = append(expressions, expr)
		}
	}
	return expressions
}

func (parser *Parser) expression() (ast.Expr, error) {
	return parser.equality()
}

func (parser *Parser) equality() (ast.Expr, error) {
	expr, err := parser.comparison()
	if err != nil {
		return nil, err
	}
	for parser.match(token.BANG_EQUAL, token.EQUAL_EQUAL) {
		operator := parser.previous()
		right, err := parser.comparison()
		if err != nil {
			return nil, err
		}
		expr = &ast.Binary{Left: expr, Operator: operator, Right: right}
	}
	return expr, nil
}

func (parser *Parser) comparison() (ast.Expr, error) {
	expr, err := parser.term()
	if err != nil {
		return nil, err
	}
	for parser.match(token.GREATER, token.GREATER_EQUAL, token.LESS, token.LESS_EQUAL) {
		operator := parser.previous()
		right, err := parser.term()
		if err != nil {
			return nil, err
		}
		expr = &ast.Binary{Left: expr, Operator: operator, Right: right}
	}
	return expr, nil
}

// Handles addition and subtraction
func (parser *Parser) term() (ast.Expr, error) {
	expr, err := parser.factor()
	if err != nil {
		return nil, err
	}
	for parser.match(token.MINUS, token.PLUS) {
		operator := parser.previous()
		right, err := parser.factor()
		if err != nil {
			return nil, err
		}
		expr = &ast.Binary{Left: expr, Operator: operator, Right: right}
	}
	return expr, nil
}

// Handles multiplication and division
func (parser *Parser) factor() (ast.Expr, error) {
	expr, err := parser.unary()
	if err != nil {
		return nil, err
	}
	for parser.match(token.SLASH, token.STAR) {
		operator := parser.previous()
		right, err := parser.unary()
		if err != nil {
			return nil, err
		}
		expr = &ast.Binary{Left: expr, Operator: operator, Right: right}
	}
	return expr, nil
}

func (parser *Parser) unary() (ast.Expr, error) {
	if parser.match(token.BANG, token.MINUS) {
		operator := parser.previous()
		right, err := parser.unary()
		if err != nil {
			return nil, err
		}
		return &ast.Unary{Operator: operator, Right: right}, nil
	}
	return parser.primary()
}

func (parser *Parser) primary() (ast.Expr, error) {
	if parser.match(token.FALSE) {
		return &ast.Literal{Value: false}, nil
	}
	if parser.match(token.TRUE) {
		return &ast.Literal{Value: true}, nil
	}
	if parser.match(token.NIL) {
		return &ast.Literal{Value: nil}, nil
	}
	if parser.match(token.NUMBER, token.STRING) {
		return &ast.Literal{Value: parser.previous().Literal}, nil
	}
	if parser.match(token.LEFT_PAREN) {
		expr, err := parser.expression()
		if err != nil {
			return nil, err
		}
		_, err = parser.consume(token.RIGHT_PAREN, "Expected ')' after expression.")
		if err != nil {
			return nil, err
		}
		return &ast.Grouping{Expression: expr}, nil
	}
	// No match!
	return nil, logger.ParserError(parser.peek(), "Expected expression.")
}

/* Internal methods */

func (parser *Parser) consume(t token.Type, message string) (token.Token, error) {
	if parser.check(t) {
		return parser.advance(), nil
	}
	return parser.previous(), logger.ParserError(parser.peek(), message)
}

func (parser *Parser) match(types ...token.Type) bool {
	for _, t := range types {
		if parser.check(t) {
			parser.advance()
			return true
		}
	}
	return false
}

func (parser *Parser) check(t token.Type) bool {
	if parser.isAtEnd() {
		return false
	}
	return parser.peek().Type == t
}

func (parser *Parser) advance() token.Token {
	if !parser.isAtEnd() {
		parser.current++
	}
	return parser.previous()
}

func (parser *Parser) isAtEnd() bool {
	return parser.peek().Type == token.EOF
}

func (parser *Parser) peek() token.Token {
	return parser.tokens[parser.current]
}

func (parser *Parser) previous() token.Token {
	return parser.tokens[parser.current-1]
}

func (parser *Parser) synchronize() {
	parser.advance()

	for !parser.isAtEnd() {
		if parser.previous().Type == token.SEMICOLON {
			return
		}

		switch parser.peek().Type {
		case token.CLASS:
		case token.FOR:
		case token.FUN:
		case token.IF:
		case token.PRINT:
		case token.RETURN:
		case token.VAR:
		case token.WHILE:
			return
		}

		parser.advance()
	}
}
