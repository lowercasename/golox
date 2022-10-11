package logger

import (
	"fmt"

	"github.com/lowercasename/golox/token"
)

var HadError = false

func LogError(line int, message string) {
	report(line, "", message)
	HadError = true
}

func report(line int, where string, message string) error {
	HadError = true
	return fmt.Errorf("[line %d] Error%v: %v\n", line, where, message)
}

func ParserError(t token.Token, message string) error {
	if t.Type == token.EOF {
		return report(t.Line, " at end", message)
	} else {
		return report(t.Line, " at '"+t.Lexeme+"'", message)
	}
}
