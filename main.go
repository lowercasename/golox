package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"github.com/lowercasename/golox/logger"
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

func run(source string) bool {
	scanner := scanner.New(source)
	tokens := scanner.ScanTokens()

	for _, token := range tokens {
		fmt.Printf("Token: %v\n", token.String())
	}

	// for tokens
	return false
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
