package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"github.com/lowercasename/golox/interpreter"
	"github.com/lowercasename/golox/logger"
	"github.com/lowercasename/golox/parser"
	"github.com/lowercasename/golox/scanner"
)

func runFile(path string) error {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	run(string(bytes))

	if logger.HadError {
		return errors.New("Scanner error!")
	}

	return nil
}

func runPrompt() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")
	for scanner.Scan() {
		run(scanner.Text())
		fmt.Print("> ")
	}
}

func run(source string) {
	scanner := scanner.New(source)
	tokens := scanner.ScanTokens()
	// for _, token := range tokens {
	// 	fmt.Printf("Token: %v\n", token.String())
	// }
	parser := parser.New(tokens)
	statements := parser.Parse()
	// if logger.HadError {
	// 	return
	// }
	// fmt.Println("==================")
	// fmt.Println("Statements:")
	// for _, statement := range statements {
	// 	fmt.Printf("%v\n", statement.String())
	// }
	// fmt.Println("==================")
	interpreter.Interpret(statements)
	return
}

func main() {
	args := os.Args[1:]
	argsCount := len(args)

	switch {
	case argsCount > 1:
		fmt.Println("Usage: golox [script]")
	case argsCount == 1:
		err := runFile(args[0])
		if err != nil {
			os.Exit(1)
		}
	default:
		runPrompt()
	}
}
