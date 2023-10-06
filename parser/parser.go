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
	var statements []ast.Expr
	for !parser.isAtEnd() {
		stmt, err := parser.declaration()
		if err != nil {
			fmt.Println(err)
			parser.synchronize()
		} else {
			statements = append(statements, stmt)
		}
	}
	return statements
}

func (parser *Parser) declaration() (ast.Expr, error) {
	if parser.match(token.FUN) {
		return parser.function("function")
	}
	if parser.match(token.VAR) {
		stmt, err := parser.varDeclaration()
		if err != nil {
			return nil, err
		}
		return stmt, err
	}
	return parser.statement()
}

func (parser *Parser) expression() (ast.Expr, error) {
	return parser.assignment()
}

func (parser *Parser) function(kind string) (ast.Stmt, error) {
	name, err := parser.consume(token.IDENTIFIER, fmt.Sprintf("Expected %s name.", kind))
	if err != nil {
		return nil, err
	}
	_, err = parser.consume(token.LEFT_PAREN, fmt.Sprintf("Expected '(' after %s name.", kind))
	if err != nil {
		return nil, err
	}
	var parameters []token.Token
	if !parser.check(token.RIGHT_PAREN) {
		for {
			if len(parameters) >= 255 {
				return nil, logger.ParserError(parser.peek(), "Cannot have more than 255 parameters.")
			}
			parameter, err := parser.consume(token.IDENTIFIER, "Expected parameter name.")
			if err != nil {
				return nil, err
			}
			parameters = append(parameters, parameter)
			if !parser.match(token.COMMA) {
				break
			}
		}
	}
	_, err = parser.consume(token.RIGHT_PAREN, "Expected ')' after parameters.")
	if err != nil {
		return nil, err
	}
	_, err = parser.consume(token.LEFT_BRACE, fmt.Sprintf("Expected '{' before %s body.", kind))
	if err != nil {
		return nil, err
	}
	body, err := parser.block()
	if err != nil {
		return nil, err
	}
	return &ast.Function{Name: name, Parameters: parameters, Body: body}, nil
}

func (parser *Parser) statement() (ast.Stmt, error) {
	if parser.match(token.PRINT) {
		stmt, err := parser.printStatement()
		if err != nil {
			return nil, err
		}
		return stmt, nil
	}
	if parser.match(token.WHILE) {
		stmt, err := parser.whileStatement()
		if err != nil {
			return nil, err
		}
		return stmt, nil
	}
	if parser.match(token.LEFT_BRACE) {
		statements, err := parser.block()
		if err != nil {
			return nil, err
		}
		// Convert slice of statements to a single block statement
		return &ast.Block{Statements: statements}, nil
	}
	if parser.match(token.FOR) {
		stmt, err := parser.forStatement()
		if err != nil {
			return nil, err
		}
		return stmt, nil
	}
	if parser.match(token.IF) {
		stmt, err := parser.ifStatement()
		if err != nil {
			return nil, err
		}
		return stmt, nil
	}
	stmt, err := parser.expressionStatement()
	if err != nil {
		return nil, err
	}
	return stmt, nil
}

func (parser *Parser) forStatement() (ast.Stmt, error) {
	_, err := parser.consume(token.LEFT_PAREN, "Expected '(' after 'for'.")
	if err != nil {
		return nil, err
	}
	var initializer ast.Stmt
	if parser.match(token.SEMICOLON) {
		initializer = nil
	} else if parser.match(token.VAR) {
		initializer, err = parser.varDeclaration()
		if err != nil {
			return nil, err
		}
	} else {
		initializer, err = parser.expressionStatement()
		if err != nil {
			return nil, err
		}
	}
	var condition ast.Expr
	if !parser.check(token.SEMICOLON) {
		condition, err = parser.expression()
		if err != nil {
			return nil, err
		}
	}
	parser.consume(token.SEMICOLON, "Expected ';' after for loop condition.")
	var increment ast.Expr
	if !parser.check(token.RIGHT_PAREN) {
		increment, err = parser.expression()
		if err != nil {
			return nil, err
		}
	}
	parser.consume(token.RIGHT_PAREN, "Expected ')' after for loop clauses.")
	body, err := parser.statement()
	if err != nil {
		return nil, err
	}
	if increment != nil {
		body = &ast.Block{Statements: []ast.Stmt{body, &ast.Expression{Expression: increment}}}
	}
	if condition != nil {
		body = &ast.While{Condition: condition, Body: body}
	}
	if initializer != nil {
		body = &ast.Block{Statements: []ast.Stmt{initializer, body}}
	}
	return body, nil
}

func (parser *Parser) ifStatement() (ast.Stmt, error) {
	_, err := parser.consume(token.LEFT_PAREN, "Expected '(' after 'if'.")
	if err != nil {
		return nil, err
	}
	condition, err := parser.expression()
	if err != nil {
		return nil, err
	}
	_, err = parser.consume(token.RIGHT_PAREN, "Expected ')' after if condition.")
	if err != nil {
		return nil, err
	}
	thenBranch, err := parser.statement()
	if err != nil {
		return nil, err
	}
	var elseBranch ast.Stmt = nil
	if parser.match(token.ELSE) {
		elseBranch, err = parser.statement()
		if err != nil {
			return nil, err
		}
	}
	return &ast.If{Condition: condition, Then: thenBranch, Else: elseBranch}, nil
}

func (parser *Parser) printStatement() (ast.Stmt, error) {
	value, err := parser.expression()
	if err != nil {
		return nil, err
	}
	_, err = parser.consume(token.SEMICOLON, "Expected ';' after value.")
	if err != nil {
		return nil, err
	}
	return &ast.Print{Expression: value}, nil
}

func (parser *Parser) varDeclaration() (ast.Stmt, error) {
	name, err := parser.consume(token.IDENTIFIER, "Expected variable name.")
	if err != nil {
		return nil, err
	}
	var initializer ast.Expr = nil
	if parser.match(token.EQUAL) {
		initializer, err = parser.expression()
		if err != nil {
			return nil, err
		}
	}
	_, err = parser.consume(token.SEMICOLON, "Expected ';' after variable declaration.")
	if err != nil {
		return nil, err
	}
	return &ast.Var{Name: name, Initializer: initializer}, nil
}

func (parser *Parser) whileStatement() (ast.Stmt, error) {
	_, err := parser.consume(token.LEFT_PAREN, "Expected '(' after 'while'.")
	if err != nil {
		return nil, err
	}
	condition, err := parser.expression()
	if err != nil {
		return nil, err
	}
	_, err = parser.consume(token.RIGHT_PAREN, "Expected ')' after while condition.")
	if err != nil {
		return nil, err
	}
	body, err := parser.statement()
	if err != nil {
		return nil, err
	}
	return &ast.While{Condition: condition, Body: body}, nil
}

func (parser *Parser) expressionStatement() (ast.Stmt, error) {
	expr, err := parser.expression()
	if err != nil {
		return nil, err
	}
	_, err = parser.consume(token.SEMICOLON, "Expected ';' after expression.")
	if err != nil {
		return nil, err
	}
	return &ast.Expression{Expression: expr}, nil
}

func (parser *Parser) block() ([]ast.Stmt, error) {
	var statements []ast.Stmt
	for !parser.check(token.RIGHT_BRACE) && !parser.isAtEnd() {
		stmt, err := parser.declaration()
		if err != nil {
			return nil, err
		}
		statements = append(statements, stmt)
	}
	_, err := parser.consume(token.RIGHT_BRACE, "Expected '}' after block.")
	if err != nil {
		return nil, err
	}
	return statements, nil
}

func (parser *Parser) assignment() (ast.Expr, error) {
	// Evaluate the l-value
	expr, err := parser.or()
	if err != nil {
		return nil, err
	}
	if parser.match(token.EQUAL) {
		// Evaluate the r-value
		equals := parser.previous()
		// Recursively evaluate the r-value
		value, err := parser.assignment()
		if err != nil {
			return nil, err
		}
		// Check if the l-value is a variable
		switch expr := expr.(type) {
		case *ast.Variable:
			return &ast.Assign{Name: expr.Name, Value: value}, nil
		}
		return nil, logger.ParserError(equals, "Invalid assignment target.")
	}
	return expr, nil
}

func (parser *Parser) or() (ast.Expr, error) {
	expr, err := parser.and()
	if err != nil {
		return nil, err
	}
	for parser.match(token.OR) {
		operator := parser.previous()
		right, err := parser.and()
		if err != nil {
			return nil, err
		}
		expr = &ast.Logical{Left: expr, Operator: operator, Right: right}
	}
	return expr, nil
}

func (parser *Parser) and() (ast.Expr, error) {
	expr, err := parser.equality()
	if err != nil {
		return nil, err
	}
	for parser.match(token.AND) {
		operator := parser.previous()
		right, err := parser.equality()
		if err != nil {
			return nil, err
		}
		expr = &ast.Logical{Left: expr, Operator: operator, Right: right}
	}
	return expr, nil
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
	return parser.call()
}

func (parser *Parser) call() (ast.Expr, error) {
	expr, err := parser.primary()
	if err != nil {
		return nil, err
	}
	for {
		if parser.match(token.LEFT_PAREN) {
			expr, err = parser.finishCall(expr)
			if err != nil {
				return nil, err
			}
		} else {
			break
		}
	}
	return expr, nil
}

func (parser *Parser) finishCall(callee ast.Expr) (ast.Expr, error) {
	var arguments []ast.Expr
	if !parser.check(token.RIGHT_PAREN) {
		for {
			if len(arguments) >= 255 {
				return nil, logger.ParserError(parser.peek(), "Cannot have more than 255 arguments.")
			}
			argument, err := parser.expression()
			if err != nil {
				return nil, err
			}
			arguments = append(arguments, argument)
			if !parser.match(token.COMMA) {
				break
			}
		}
	}
	paren, err := parser.consume(token.RIGHT_PAREN, "Expected ')' after function arguments.")
	if err != nil {
		return nil, err
	}
	return &ast.Call{Callee: callee, Paren: paren, Arguments: arguments}, nil
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
	if parser.match(token.IDENTIFIER) {
		return &ast.Variable{Name: parser.previous()}, nil
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
