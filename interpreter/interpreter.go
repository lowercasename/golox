package interpreter

import (
	"fmt"
	"time"

	"github.com/lowercasename/golox/ast"
	"github.com/lowercasename/golox/environment"
	"github.com/lowercasename/golox/logger"
	"github.com/lowercasename/golox/token"
)

type Interpreter struct {
	environment *environment.Environment
}

type Callable interface {
	Call(interpreter *Interpreter, arguments []any) (any, error)
	Arity() int
}

type Function struct {
	Callable
	declaration *ast.Function
}

type NativeFunction struct {
	Callable
	nativeCall func(interpreter *Interpreter, arguments []any) (any, error)
	arity      int
}

func (f NativeFunction) Arity() int {
	return f.arity
}

func (f NativeFunction) Call(interpreter *Interpreter, arguments []any) (any, error) {
	return f.nativeCall(interpreter, arguments)
}

func (f Function) Arity() int {
	return len(f.declaration.Parameters)
}

func (f Function) Call(interpreter *Interpreter, arguments []any) (any, error) {
	interpreter.environment = environment.NewEnclosed(interpreter.environment)
	for i, param := range f.declaration.Parameters {
		interpreter.environment.Define(param.Lexeme, arguments[i])
	}
	for _, statement := range f.declaration.Body {
		_, err := interpreter.evaluate(statement)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func New() *Interpreter {
	globals := environment.New()
	globals.Define("clock", NativeFunction{
		nativeCall: func(interpreter *Interpreter, arguments []any) (any, error) {
			// Return time in seconds
			return int(time.Now().UnixMilli()) / 1000, nil
		},
		arity: 0,
	})
	globals.Define("sqrt", NativeFunction{
		nativeCall: func(interpreter *Interpreter, arguments []any) (any, error) {
			argument := arguments[0].(float64)
			return float64(argument * argument), nil
		},
		arity: 1,
	})
	return &Interpreter{
		environment: globals,
	}
}

func (i *Interpreter) Interpret(expressions []ast.Expr) {
	for _, expr := range expressions {
		_, err := i.evaluate(expr)
		if err != nil {
			fmt.Print(err)
			return
		}
	}
}

func (i *Interpreter) evaluate(expr ast.Expr) (any, error) {
	switch expr.(type) {
	case *ast.Literal:
		v, err := i.literal(expr)
		if err != nil {
			return nil, err
		}
		return v, nil
	case *ast.Grouping:
		v, err := i.grouping(expr)
		if err != nil {
			return nil, err
		}
		return v, nil
	case *ast.Unary:
		v, err := i.unary(expr)
		if err != nil {
			return nil, err
		}
		return v, nil
	case *ast.Binary:
		v, err := i.binary(expr)
		if err != nil {
			return nil, err
		}
		return v, nil
	case *ast.Var:
		v, err := i.variableStmt(expr)
		if err != nil {
			return nil, err
		}
		return v, nil
	case *ast.Assign:
		v, err := i.assign(expr)
		if err != nil {
			return nil, err
		}
		return v, nil
	case *ast.Print:
		_, err := i.print(expr)
		if err != nil {
			return nil, err
		}
		return nil, nil
	case *ast.Expression:
		v, err := i.evaluate(expr.(*ast.Expression).Expression)
		if err != nil {
			return nil, err
		}
		return v, nil
	case *ast.Variable:
		v, err := i.variableExpr(expr)
		if err != nil {
			return nil, err
		}
		return v, nil
	case *ast.Block:
		v, err := i.block(expr)
		if err != nil {
			return nil, err
		}
		return v, nil
	case *ast.If:
		v, err := i.ifStmt(expr)
		if err != nil {
			return nil, err
		}
		return v, nil
	case *ast.Logical:
		v, err := i.logical(expr)
		if err != nil {
			return nil, err
		}
		return v, nil
	case *ast.While:
		v, err := i.whileStmt(expr)
		if err != nil {
			return nil, err
		}
		return v, nil
	case *ast.Call:
		v, err := i.evaluate(expr.(*ast.Call).Callee)
		if err != nil {
			return nil, err
		}
		arguments := expr.(*ast.Call).Arguments
		// Evaluate the arguments.
		var evaluatedArguments []any
		for _, argument := range arguments {
			argument, err := i.evaluate(argument)
			if err != nil {
				return nil, err
			}
			evaluatedArguments = append(evaluatedArguments, argument)
		}
		// Get the function from the callee.
		c, ok := v.(Callable)
		if !ok {
			return nil, logger.InterpreterError("Can only call functions and classes.")
		}
		if len(evaluatedArguments) != c.Arity() {
			return nil, logger.InterpreterError(fmt.Sprintf("Expected %d arguments but got %d.", c.Arity(), len(evaluatedArguments)))
		}
		return c.Call(i, evaluatedArguments)
	case *ast.Function:
		function := Function{declaration: expr.(*ast.Function)}
		i.environment.Define(function.declaration.Name.Lexeme, function)
		return nil, nil
	}
	return nil, logger.InterpreterError("Unknown expression type: " + fmt.Sprintf("%T", expr))
}

func (i *Interpreter) block(expr ast.Expr) (any, error) {
	// Save the current environment so we can restore it later.
	previousEnvironment := i.environment
	// Create a new environment for the block.
	i.environment = environment.NewEnclosed(previousEnvironment)
	for _, statement := range expr.(*ast.Block).Statements {
		_, err := i.evaluate(statement)
		if err != nil {
			// Restore the previous environment before returning the error
			i.environment = previousEnvironment
			return nil, err
		}
	}
	// Restore the previous environment.
	i.environment = previousEnvironment
	return nil, nil
}

func (i *Interpreter) literal(expr ast.Expr) (any, error) {
	v := expr.(*ast.Literal).Value
	return v, nil
}

func (i *Interpreter) logical(expr ast.Expr) (any, error) {
	logicalExpr := expr.(*ast.Logical)
	// Evaluate the left operand first.
	left, err := i.evaluate(logicalExpr.Left)
	if err != nil {
		return nil, err
	}
	if logicalExpr.Operator.Type == token.OR {
		// If the left operand is true and we're doing an OR, we can short-circuit and return it.
		if isTruthy(left) {
			return left, nil
		}
	} else if logicalExpr.Operator.Type == token.AND {
		// If the left operand is false and we're doing an AND, we can short-circuit and return it.
		if !isTruthy(left) {
			return left, nil
		}
	}
	return i.evaluate(logicalExpr.Right)
}

func (i *Interpreter) grouping(expr ast.Expr) (any, error) {
	grouping := expr.(*ast.Grouping)
	v, err := i.evaluate(grouping.Expression)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (i *Interpreter) unary(expr ast.Expr) (any, error) {
	unary := expr.(*ast.Unary)
	right, err := i.evaluate(unary.Right)
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
	return nil, logger.InterpreterError("Unknown unary operator.")
}

func (i *Interpreter) binary(expr ast.Expr) (any, error) {
	binary := expr.(*ast.Binary)
	left, err := i.evaluate(binary.Left)
	if err != nil {
		return nil, err
	}
	right, err := i.evaluate(binary.Right)
	if err != nil {
		return nil, err
	}
	switch binary.Operator.Type {
	case token.MINUS:
		err := checkNumberOperands(binary.Operator, left, right)
		if err != nil {
			return nil, err
		}
		return left.(float64) - right.(float64), nil
	case token.SLASH:
		err := checkNumberOperands(binary.Operator, left, right)
		if err != nil {
			return nil, err
		}
		// Check for division by zero.
		if right.(float64) == 0 {
			return nil, logger.InterpreterErrorWithLineNumber(binary.Operator, "Division by zero. Eldritch horrors invoked.")
		}
		return left.(float64) / right.(float64), nil
	case token.STAR:
		err := checkNumberOperands(binary.Operator, left, right)
		if err != nil {
			return nil, err
		}
		return left.(float64) * right.(float64), nil
	case token.PLUS:
		switch leftTerm := left.(type) {
		case float64:
			switch rightTerm := right.(type) {
			case float64:
				return leftTerm + rightTerm, nil
			case string:
				// If the left term is a number and the right term is a string, convert the number to a string and concatenate.
				return fmt.Sprintf("%v%v", leftTerm, rightTerm), nil
			}
		case string:
			switch rightTerm := right.(type) {
			case float64:
				// If the left term is a string and the right term is a number, convert the number to a string and concatenate.
				return fmt.Sprintf("%v%v", leftTerm, rightTerm), nil
			case string:
				return leftTerm + rightTerm, nil
			}
		}
		return nil, logger.InterpreterErrorWithLineNumber(binary.Operator, "Operands of '+' must both be either numbers or strings.")
	case token.GREATER:
		err := checkNumberOperands(binary.Operator, left, right)
		if err != nil {
			return nil, err
		}
		return left.(float64) > right.(float64), nil
	case token.GREATER_EQUAL:
		err := checkNumberOperands(binary.Operator, left, right)
		if err != nil {
			return nil, err
		}
		return left.(float64) >= right.(float64), nil
	case token.LESS:
		err := checkNumberOperands(binary.Operator, left, right)
		if err != nil {
			return nil, err
		}
		return left.(float64) < right.(float64), nil
	case token.LESS_EQUAL:
		err := checkNumberOperands(binary.Operator, left, right)
		if err != nil {
			return nil, err
		}
		return left.(float64) <= right.(float64), nil
	case token.BANG_EQUAL:
		return !isEqual(left, right), nil
	case token.EQUAL_EQUAL:
		return isEqual(left, right), nil
	}
	return nil, logger.InterpreterError("Evaluation failed.")
}

func (i *Interpreter) ifStmt(expr ast.Expr) (any, error) {
	ifStmt := expr.(*ast.If)
	condition, err := i.evaluate(ifStmt.Condition)
	if err != nil {
		return nil, err
	}
	if isTruthy(condition) {
		return i.evaluate(ifStmt.Then)
	} else if ifStmt.Else != nil {
		return i.evaluate(ifStmt.Else)
	}
	return nil, nil
}

func (i *Interpreter) print(expr ast.Expr) (any, error) {
	print := expr.(*ast.Print)
	v, err := i.evaluate(print.Expression)
	if err != nil {
		return nil, err
	}
	fmt.Println(v)
	return nil, nil
}

// Declare a variable in the current scope.
func (i *Interpreter) variableStmt(expr ast.Expr) (any, error) {
	variableStmt := expr.(*ast.Var)
	var v any = nil
	var err error
	// If the variable has an initializer, evaluate it.
	if variableStmt.Initializer != nil {
		v, err = i.evaluate(variableStmt.Initializer)
		if err != nil {
			return nil, err
		}
	}
	// Declare the variable. If it wasn't initialized, it will be nil.
	i.environment.Define(variableStmt.Name.Lexeme, v)
	return nil, nil
}

func (i *Interpreter) whileStmt(expr ast.Expr) (any, error) {
	whileStmt := expr.(*ast.While)
	for {
		// Evaluate the condition.
		condition, err := i.evaluate(whileStmt.Condition)
		if err != nil {
			return nil, err
		}
		// If the condition is false, break out of the loop.
		if !isTruthy(condition) {
			break
		}
		// Evaluate the body.
		_, err = i.evaluate(whileStmt.Body)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

// Assign a value to a variable.
func (i *Interpreter) assign(expr ast.Expr) (any, error) {
	assign := expr.(*ast.Assign)
	// Evaluate the value to assign because otherwise it will end up
	// as a pointer to ast.Literal and not the actual value.
	v, err := i.evaluate(assign.Value)
	if err != nil {
		return nil, err
	}
	_, err2 := i.environment.Assign(assign.Name, v)
	if err2 != nil {
		return nil, err2
	}
	return v, nil
}

func (i *Interpreter) variableExpr(expr ast.Expr) (any, error) {
	variableExpr := expr.(*ast.Variable)
	v, err := i.environment.Get(variableExpr.Name)
	if err != nil {
		return nil, err
	}
	return v, nil
}

/* Helper functions */

func isTruthy(value any) bool {
	// nil is falsey.
	if value == nil {
		return false
	}
	// Booleans are truthy or falsey.
	if value, ok := value.(bool); ok {
		return value
	}
	// Everything else is truthy.
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
	return logger.InterpreterErrorWithLineNumber(operator, "Operand must be a number.")
}

func checkNumberOperands(operator token.Token, left any, right any) error {
	switch left.(type) {
	case int, float64:
		switch right.(type) {
		case int, float64:
			return nil
		}
		return logger.InterpreterErrorWithLineNumber(operator, "Right operand must be a number.")
	}
	return logger.InterpreterErrorWithLineNumber(operator, "Left operand must be a number.")
}

func stringify(value any) string {
	if value == nil {
		return "nil"
	}
	return fmt.Sprintf("%v", value)
}
