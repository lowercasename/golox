package environment

import (
	"github.com/lowercasename/golox/logger"
	"github.com/lowercasename/golox/token"
)

type Environment struct {
	Enclosing *Environment
	Values    map[string]any
}

func New() *Environment {
	return &Environment{
		Values: make(map[string]any),
	}
}

func NewEnclosed(enclosing *Environment) *Environment {
	env := New()
	env.Enclosing = enclosing
	return env
}

func (e *Environment) Define(name string, value any) {
	e.Values[name] = value
}

func (e *Environment) Get(name token.Token) (any, error) {
	// First, check the current environment
	if value, ok := e.Values[name.Lexeme]; ok {
		// If the variable is set to nil, it hasn't been initialized yet - this is a runtime error
		if value == nil {
			return nil, logger.InterpreterErrorWithLineNumber(name, "Variable '"+name.Lexeme+"' used before being initialized.")
		}
		return value, nil
	}
	// If a value is not found there, check the enclosing environment
	if e.Enclosing != nil {
		return e.Enclosing.Get(name)
	}
	// Otherwise, error
	return nil, logger.InterpreterErrorWithLineNumber(name, "Undefined variable '"+name.Lexeme+"'.")
}

func (e *Environment) Assign(name token.Token, value any) (any, error) {
	// If current environment contains the variable, assign it
	if _, ok := e.Values[name.Lexeme]; ok {
		e.Values[name.Lexeme] = value
		return value, nil
	}
	// If not, check the enclosing environment
	if e.Enclosing != nil {
		return e.Enclosing.Assign(name, value)
	}
	// Otherwise, error
	return nil, logger.InterpreterErrorWithLineNumber(name, "Undefined variable '"+name.Lexeme+"'.")
}
