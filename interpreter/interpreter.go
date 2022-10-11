package interpreter

import (
	"errors"
	"fmt"

	"github.com/lowercasename/golox/ast"
	"github.com/lowercasename/golox/token"
)

func Interpret(expressions []ast.Expr) {
	for _, expr := range expressions {
		v, err := evaluate(expr)
		if err != nil {
			panic(err)
		}
		fmt.Println(stringify(v))
	}
}

func evaluate(expr ast.Expr) (any, error) {
	switch expr.(type) {
	case *ast.Literal:
		v, err := literal(expr)
		if err != nil {
			return nil, err
		}
		return v, nil
	case *ast.Grouping:
		v, err := grouping(expr)
		if err != nil {
			return nil, err
		}
		return v, nil
	case *ast.Unary:
		v, err := unary(expr)
		if err != nil {
			return nil, err
		}
		return v, nil
	case *ast.Binary:
		v, err := binary(expr)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
	return nil, errors.New("Evaluation failed.")
}

func literal(n ast.Expr) (any, error) {
	v := n.(*ast.Literal).Value
	return v, nil
}

func grouping(n ast.Expr) (any, error) {
	grouping := n.(*ast.Grouping)
	v, err := evaluate(grouping.Expression)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func unary(n ast.Expr) (any, error) {
	unary := n.(*ast.Unary)
	right, err := evaluate(unary.Right)
	if err != nil {
		return nil, err
	}
	switch unary.Operator.Type {
	case token.MINUS:
		err := checkNumberOperand(unary.Operator, right)
		if err != nil {
			return nil, err
		}
		return -right.(float64), nil
	case token.BANG:
		return !isTruthy(right), nil
	}
	return nil, errors.New("Evaluation failed.")
}

func binary(n ast.Expr) (any, error) {
	binary := n.(*ast.Binary)
	left, err := evaluate(binary.Left)
	if err != nil {
		return nil, err
	}
	right, err := evaluate(binary.Right)
	if err != nil {
		return nil, err
	}
	switch binary.Operator.Type {
	case token.MINUS:
		checkNumberOperands(binary.Operator, left, right)
		return left.(float64) - right.(float64), nil
	case token.SLASH:
		checkNumberOperands(binary.Operator, left, right)
		return left.(float64) / right.(float64), nil
	case token.STAR:
		checkNumberOperands(binary.Operator, left, right)
		return left.(float64) * right.(float64), nil
	case token.PLUS:
		switch leftTerm := left.(type) {
		case float64:
			switch rightTerm := right.(type) {
			case float64:
				return leftTerm + rightTerm, nil
			}
		case string:
			switch rightTerm := right.(type) {
			case string:
				return leftTerm + rightTerm, nil
			}
		}
		return nil, errors.New("Operands of '+' must both be either numbers or strings.")
	case token.GREATER:
		checkNumberOperands(binary.Operator, left, right)
		return left.(float64) > right.(float64), nil
	case token.GREATER_EQUAL:
		checkNumberOperands(binary.Operator, left, right)
		return left.(float64) >= right.(float64), nil
	case token.LESS:
		checkNumberOperands(binary.Operator, left, right)
		return left.(float64) < right.(float64), nil
	case token.LESS_EQUAL:
		checkNumberOperands(binary.Operator, left, right)
		return left.(float64) <= right.(float64), nil
	case token.BANG_EQUAL:
		return !isEqual(left, right), nil
	case token.EQUAL_EQUAL:
		return isEqual(left, right), nil
	}
	return nil, errors.New("Evaluation failed.")
}

func isTruthy(value any) bool {
	if value == nil {
		return false
	}
	v := value.(bool)
	if v {
		return v
	}
	return true
}

func isEqual(a any, b any) bool {
	// Nil is only equal to nil.
	if a == nil && b == nil {
		return true
	}
	// If one is nil and the other isn't, they're not equal.
	if a == nil {
		return false
	}
	// If they're both numbers, compare them.
	return a == b
}

func checkNumberOperand(operator token.Token, operand any) error {
	switch operand.(type) {
	case int, float64:
		return nil
	}
	return errors.New("Operand must be a number.")
}

func checkNumberOperands(operator token.Token, left any, right any) error {
	switch left.(type) {
	case int, float64:
		switch right.(type) {
		case int, float64:
			return nil
		}
		return errors.New("Right operand must be a number.")
	}
	return errors.New("Left operand must be a number.")
}

func stringify(value any) string {
	if value == nil {
		return "nil"
	}
	return fmt.Sprintf("%v", value)
}
