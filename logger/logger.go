package logger

import (
	"fmt"

	"github.com/lowercasename/golox/token"
)

var HadError = false

func ScannerError(line int, message string) error {
	return report(line, "", message, "Scanner")
}

func report(line int, where string, message string, errorType string) error {
	HadError = true
	return fmt.Errorf("[line %d] %vError%v: %v\n", line, errorType, where, message)
}

func ParserError(t token.Token, message string) error {
	if t.Type == token.EOF {
		return report(t.Line, " at end", message, "Parser")
	} else {
		return report(t.Line, " at '"+t.Lexeme+"'", message, "Parser")
	}
}

func InterpreterError(message string) error {
	return fmt.Errorf("Error: %v\n", message)
}

func InterpreterErrorWithLineNumber(t token.Token, message string) error {
	return report(t.Line, " at '"+t.Lexeme+"'", message, "Runtime")
}
