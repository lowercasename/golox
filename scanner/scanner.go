package scanner

import (
	"fmt"
	"strconv"

	"github.com/lowercasename/golox/logger"
	"github.com/lowercasename/golox/token"
)

var keywords = map[string]token.Type{
	"and":    token.AND,
	"class":  token.CLASS,
	"else":   token.ELSE,
	"false":  token.FALSE,
	"for":    token.FOR,
	"fun":    token.FUN,
	"if":     token.IF,
	"nil":    token.NIL,
	"or":     token.OR,
	"print":  token.PRINT,
	"return": token.RETURN,
	"super":  token.SUPER,
	"this":   token.THIS,
	"true":   token.TRUE,
	"var":    token.VAR,
	"while":  token.WHILE,
}

type Scanner struct {
	source  string
	start   int
	current int
	line    int
	tokens  []token.Token
}

// Creates a new scanner
func New(source string) Scanner {
	scanner := Scanner{source: source, line: 1, tokens: make([]token.Token, 0)}
	return scanner
}

func (scanner *Scanner) ScanTokens() []token.Token {
	for !scanner.isAtEnd() {
		// We're at the beginning of the next lexeme
		scanner.start = scanner.current
		scanner.scanToken()
	}
	// Add an EOF after all other tokens
	scanner.tokens = append(scanner.tokens, token.Token{Type: token.EOF, Lexeme: "", Literal: nil, Line: scanner.line})
	return scanner.tokens
}

func (scanner *Scanner) addToken(tokenType token.Type, literal any) {
	text := scanner.source[scanner.start:scanner.current]
	scanner.tokens = append(scanner.tokens, token.Token{Type: tokenType, Lexeme: text, Literal: literal, Line: scanner.line})
}

func (scanner *Scanner) handleIdentifier() {
	for scanner.isAlphaNumeric(scanner.peek()) {
		scanner.current++
	}
	tokenString := string(scanner.source[scanner.start:scanner.current])
	// Check if the identifier is a reserved keyword
	tokenType, identifierIsReservedKeyword := keywords[tokenString]
	if identifierIsReservedKeyword {
		scanner.addToken(tokenType, nil)
	} else {
		scanner.addToken(token.IDENTIFIER, nil)
	}
}

func (scanner *Scanner) handleString() {
	// Keep advancing to closing ", including over newlines
	for scanner.peek() != '"' && !scanner.isAtEnd() {
		if scanner.peek() == '\n' {
			scanner.line++
		}
		scanner.current++
	}
	// Unterminated string
	if scanner.isAtEnd() {
		fmt.Printf(logger.ScannerError(scanner.line, "Unterminated string.").Error())
		return
	}
	// Consume the closing "
	scanner.current++
	// Trim the surrounding quotes
	stringValue := string(scanner.source[scanner.start+1 : scanner.current-1])
	scanner.addToken(token.STRING, stringValue)
}

func (scanner *Scanner) handleNumber() {
	for scanner.isDigit(scanner.peek()) {
		scanner.current++
	}
	// Look for a fractional part
	if scanner.peek() == '.' && scanner.isDigit(scanner.peekNext()) {
		// Consume the "."
		scanner.current++
		for scanner.isDigit(scanner.peek()) {
			scanner.current++
		}
	}
	numString := string(scanner.source[scanner.start:scanner.current])
	numValue, err := strconv.ParseFloat(numString, 64)
	if err != nil {
		fmt.Printf(logger.ScannerError(scanner.line, "Could not convert number literal to float.").Error())
		return
	}
	scanner.addToken(token.NUMBER, numValue)
}

func (scanner *Scanner) scanToken() {
	// Move to the next character (byte) of the source
	c := scanner.advance()

	switch c {
	case '(':
		scanner.addToken(token.LEFT_PAREN, nil)
	case ')':
		scanner.addToken(token.RIGHT_PAREN, nil)
	case '{':
		scanner.addToken(token.LEFT_BRACE, nil)
	case '}':
		scanner.addToken(token.RIGHT_BRACE, nil)
	case ',':
		scanner.addToken(token.COMMA, nil)
	case '.':
		scanner.addToken(token.DOT, nil)
	case '-':
		scanner.addToken(token.MINUS, nil)
	case '+':
		scanner.addToken(token.PLUS, nil)
	case ';':
		scanner.addToken(token.SEMICOLON, nil)
	case '*':
		scanner.addToken(token.STAR, nil)
	case '!':
		if scanner.match('=') {
			scanner.addToken(token.BANG_EQUAL, nil)
		} else {
			scanner.addToken(token.BANG, nil)
		}
	case '=':
		if scanner.match('=') {
			scanner.addToken(token.EQUAL_EQUAL, nil)
		} else {
			scanner.addToken(token.EQUAL, nil)
		}
	case '<':
		if scanner.match('=') {
			scanner.addToken(token.LESS_EQUAL, nil)
		} else {
			scanner.addToken(token.LESS, nil)
		}
	case '>':
		if scanner.match('=') {
			scanner.addToken(token.GREATER_EQUAL, nil)
		} else {
			scanner.addToken(token.GREATER, nil)
		}
	case '/':
		// If we have two forward slashes, this is a comment
		if scanner.match('/') {
			// Keep advancing to end of comment line
			for scanner.peek() != '\n' && !scanner.isAtEnd() {
				scanner.current++
			}
		} else if scanner.match('*') {
			// If we have a forward slash and an asterisk, this is a block comment
			// Keep advancing to end of comment block
			for !(scanner.peek() == '*' && scanner.peekNext() == '/') && !scanner.isAtEnd() {
				if scanner.peek() == '\n' {
					scanner.line++
				}
				scanner.current++
			}
			// Unterminated comment block
			if scanner.isAtEnd() {
				fmt.Printf(logger.ScannerError(scanner.line, "Unterminated comment block.").Error())
				return
			}
			// Consume the closing */
			scanner.current += 2
		} else {
			scanner.addToken(token.SLASH, nil)
		}
	case ' ', '\r', '\t':
		// Ignore whitespace
	case '\n':
		scanner.line++
	case '"':
		scanner.handleString()
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		scanner.handleNumber()
	default:
		// At this point it's either an identifier or an unexpected
		// character.
		if scanner.isAlpha(c) {
			scanner.handleIdentifier()
		} else {
			fmt.Printf(logger.ScannerError(scanner.line, "Unexpected charater.").Error())
		}
	}
}

func (scanner *Scanner) isAtEnd() bool {
	return scanner.current >= len(scanner.source)
}

func (scanner *Scanner) isDigit(b byte) bool {
	return b >= 0x30 && b <= 0x39
}

func (scanner *Scanner) isAlpha(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_'
}

func (scanner *Scanner) isAlphaNumeric(b byte) bool {
	return scanner.isAlpha(b) || scanner.isDigit(b)
}

// advance returns the current character and advances to the next
func (sc *Scanner) advance() byte {
	sc.current++
	return sc.source[sc.current-1]
}

func (scanner *Scanner) match(expected byte) bool {
	if scanner.isAtEnd() {
		return false
	}
	if scanner.source[scanner.current] != expected {
		return false
	}
	scanner.current++
	return true
}

func (scanner *Scanner) peek() byte {
	if scanner.isAtEnd() {
		return 0
	}
	return scanner.source[scanner.current]
}

func (scanner *Scanner) peekNext() byte {
	if scanner.current+1 >= len(scanner.source) {
		return 0
	}
	return scanner.source[scanner.current+1]
}
