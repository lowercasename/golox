package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/lowercasename/golox/interpreter"
	"github.com/lowercasename/golox/parser"
	"github.com/lowercasename/golox/scanner"
	"github.com/pkg/term"
)

const (
	version = "0.1.0"
)

// Raw input keycodes
var up byte = 65
var down byte = 66
var right byte = 67
var left byte = 68
var escape byte = 27
var enter byte = 13
var delete byte = 127
var backspace byte = 8
var ctrlC byte = 3
var ctrlD byte = 4
var keys = map[byte]bool{
	up:    true,
	down:  true,
	right: true,
	left:  true,
}

func runFile(path string, debug bool) error {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	interpreter := interpreter.New()
	run(string(bytes), interpreter, debug)
	return nil
}

func runPrompt(debug bool) {
	scanner := bufio.NewScanner(os.Stdin)
	interpreter := interpreter.New()
	fmt.Print("> ")
	for scanner.Scan() {
		run(scanner.Text(), interpreter, debug)
		fmt.Print("> ")
	}
}

// getInput will read raw input from the terminal
// It returns the raw ASCII value inputted
// From: https://github.com/Nexidian/gocliselect
func getInput() byte {
	t, _ := term.Open("/dev/tty")

	err := term.RawMode(t)
	if err != nil {
		log.Fatal(err)
	}

	var readBytesNumber int
	readBytes := make([]byte, 3)
	readBytesNumber, err = t.Read(readBytes)

	t.Restore()
	t.Close()

	// Arrow keys are prefixed with the ANSI escape code which take up the first two bytes.
	// The third byte is the key specific value we are looking for.
	// For example the up arrow key is '<esc>[A' while the right is '<esc>[C'
	// See: https://en.wikipedia.org/wiki/ANSI_escape_code
	if readBytesNumber == 3 {
		if _, ok := keys[readBytes[2]]; ok {
			return readBytes[2]
		}
	} else {
		return readBytes[0]
	}

	return 0
}

func runRawPrompt(debug bool) string {
	fmt.Println("Welcome to Golox " + version + "!")
	fmt.Println("Press Ctrl+C or Ctrl+D to exit.")
	// Print the prompt
	fmt.Print("> ")
	interpreter := interpreter.New()
	currentInput := ""
	// Set up a command history
	history := []string{}
	// Set up a pointer to the current command in the history
	historyPointer := 0
	// Set up a pointer to the current position in the current command
	positionPointer := 0
	for {
		keyCode := getInput()
		if keyCode == ctrlC || keyCode == ctrlD || keyCode == escape {
			fmt.Println("\nBye!")
			os.Exit(0)
		} else if keyCode == delete || keyCode == backspace {
			// Delete the character at the current position
			if positionPointer > 0 {
				currentInput = currentInput[:positionPointer-1] + currentInput[positionPointer:]
				positionPointer--
				// Erase the current line
				fmt.Print("\033[2K\r")
				// Print the current input
				fmt.Print("\r> " + currentInput)
				// Move the cursor back to the current position
				for i := 0; i < len(currentInput)-positionPointer; i++ {
					fmt.Print("\033[1D")
				}
			}
		} else if keyCode == enter {
			// Print a newline to the terminal
			fmt.Print("\n")
			// DEBUG: Print the current input
			if debug {
				fmt.Println("DEBUG: " + currentInput)
			}
			// Send input to interpreter
			run(currentInput, interpreter, debug)
			// Add input to history
			history = append(history, currentInput)
			// Reset the history pointer
			historyPointer = len(history)
			// Reset the position pointer
			positionPointer = 0
			// Clear the current input
			currentInput = ""
			// Print the prompt
			fmt.Print("\r> ")
		} else if keyCode == up {
			// If the history pointer is not at the beginning of the history
			if historyPointer > 0 {
				// Move the history pointer back one
				historyPointer--
				// Remove the current input from the terminal
				for i := 0; i < len(currentInput); i++ {
					fmt.Print("\b \b")
				}
				// Print the prompt
				fmt.Print("\r> ")
				// Print the command fetched from the history
				fmt.Print(history[historyPointer])
				// Set the current input to the command fetched from the history
				currentInput = history[historyPointer]
				// Set the position pointer to the end of the current input
				positionPointer = len(currentInput)
			}
		} else if keyCode == down {
			// If the history pointer is not at the end of the history
			if historyPointer < len(history)-1 {
				// Move the history pointer forward one
				historyPointer++
				// Remove the current input from the terminal
				for i := 0; i < len(currentInput); i++ {
					fmt.Print("\b \b")
				}
				// Print the prompt
				fmt.Print("\r> ")
				// Print the command fetched from the history
				fmt.Print(history[historyPointer])
				// Set the current input to the command fetched from the history
				currentInput = history[historyPointer]
				// Set the position pointer to the end of the current input
				positionPointer = len(currentInput)
			} else {
				// If the history pointer is at the end of the history
				// Remove the current input from the terminal
				for i := 0; i < len(currentInput); i++ {
					fmt.Print("\b \b")
				}
				// Print the prompt
				fmt.Print("\r> ")
				// Reset the current input
				currentInput = ""
				// Reset the position pointer
				positionPointer = 0
				// Set the history pointer to the end of the history
				historyPointer = len(history)
			}
		} else if keyCode == left {
			// If the position pointer is not at the beginning of the current input
			if positionPointer > 0 {
				// Move the position pointer back one
				positionPointer--
				// Move the cursor back by escape code
				fmt.Print("\033[1D")
			}
		} else if keyCode == right {
			// If the position pointer is not at the end of the current input
			if positionPointer < len(currentInput) {
				// Move the position pointer forward one
				positionPointer++
				// Move the cursor forward
				fmt.Print("\033[1C")
			}
		} else if keyCode >= 32 && keyCode <= 126 { // Printable ASCII characters
			// Insert the character at the current position
			currentInput = currentInput[:positionPointer] + string(keyCode) + currentInput[positionPointer:]
			positionPointer++
			// Erase the current line
			fmt.Print("\033[2K\r")
			// Print the current input
			fmt.Print("\r> " + currentInput)
			// Move the cursor back to the current position
			for i := 0; i < len(currentInput)-positionPointer; i++ {
				fmt.Print("\033[1D")
			}
		}
	}
}

func run(source string, interpreter *interpreter.Interpreter, debug bool) {
	scanner := scanner.New(source)
	tokens := scanner.ScanTokens()
	if debug {
		fmt.Println("==================")
		fmt.Println("Tokens:")
		for _, token := range tokens {
			fmt.Printf("Token: %v\n", token.String())
		}
		fmt.Println("==================")
	}
	parser := parser.New(tokens)
	statements := parser.Parse()
	if debug {
		fmt.Println("==================")
		fmt.Println("Statements:")
		for _, statement := range statements {
			fmt.Printf("%v\n", statement.String())
		}
		fmt.Println("==================")
	}
	interpreter.Interpret(statements)
	return
}

func main() {
	args := os.Args[1:]
	argsCount := len(args)

	switch {
	case argsCount > 2:
		fmt.Println("Usage: golox [script] [--debug]")
	case argsCount == 1:
		if args[0] == "--debug" {
			runRawPrompt(true)
		} else {
			err := runFile(args[0], false)
			if err != nil {
				fmt.Println(err)
			}
		}
	case argsCount == 2:
		if args[1] == "--debug" {
			runFile(args[0], true)
		} else {
			fmt.Println("Usage: golox [script] [--debug]")
		}
	default:
		runRawPrompt(false)
	}
}
