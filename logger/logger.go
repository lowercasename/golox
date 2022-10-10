package logger

import "fmt"

var HadError = false

func LogError(line int, message string) {
	report(line, "", message)
	HadError = true
}

func report(line int, where string, message string) {
	HadError = true
	fmt.Printf("[line %d] Error%v: %v\n", line, where, message)
}
