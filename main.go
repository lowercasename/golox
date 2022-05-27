package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"github.com/lowercasename/golox/errorreport"
	"github.com/lowercasename/golox/scanner"
	// "text/scanner"
	// "github.com/lowercasename/golox/tok"
)

func main() {
	args := os.Args[1:]
	argsCount := len(args)

	switch {
	case argsCount > 1:
		fmt.Println("Usage: jlox [script]")
	case argsCount == 1:
		err := runFile(args[0])
		if err != nil {
			os.Exit(1)
		}
	default:
		runPrompt()
	}
}

func runFile(path string) error {
	errorReport := errorreport.NewErrorReport()

	bytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	run(string(bytes), &errorReport)

	if errorReport.HadError {
		return errors.New("Scanner error!")
	}

	return nil
}

func runPrompt() {
	errorReport := errorreport.NewErrorReport()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")
	for scanner.Scan() {
		run(scanner.Text(), &errorReport)
		errorReport.HadError = false
		fmt.Print("> ")
	}
}

func run(source string, errorReport *errorreport.ErrorReport) bool {
	tokens := scanner.ScanTokens(source, errorReport)

	for _, token := range tokens {
		fmt.Printf("Token: %v\n", token)
	}

	// for tokens
	return false
}

// func reportError(line int, message string) {
//     report(line, "", message)
// }

// func report(line int, where string, message string) {
//     fmt.Printf("[line %d] Error %v: %v\n", line, where, message)
// }
