package scanner

import (
	"github.com/lowercasename/golox/errorreport"
	"github.com/lowercasename/golox/tok"
	"strconv"
)

type sourceLocation struct {
	Start   int
	Current int
	Line    int
}

func (location *sourceLocation) isAtEnd(runes []rune) bool {
	return location.Current >= len(runes)
}

func (location *sourceLocation) beginNewLexeme() {
	location.Start = location.Current
}

var keywords = map[string]tok.TokenType{
	"and":    tok.And,
	"class":  tok.Class,
	"else":   tok.Else,
	"false":  tok.False,
	"for":    tok.For,
	"fun":    tok.Fun,
	"if":     tok.If,
	"nil":    tok.Nil,
	"or":     tok.Or,
	"print":  tok.Print,
	"return": tok.Return,
	"super":  tok.Super,
	"this":   tok.This,
	"true":   tok.True,
	"var":    tok.Var,
	"while":  tok.While,
}

func ScanTokens(source string, errorReport *errorreport.ErrorReport) []tok.Token {
	// Set initial location
	location := sourceLocation{Line: 1}
	runes := []rune(source)

	// Set up tokens array
	tokens := make([]tok.Token, 0, len(runes)/2)

	for !location.isAtEnd(runes) {
		// We are at the beginning of the next lexeme
		location.beginNewLexeme()
		scanToken(&location, runes, &tokens, errorReport)
	}

	// Add an EOF after all other tokens
	addToken(&tokens, tok.EOF, nil)

	return tokens
}

func scanToken(location *sourceLocation, runes []rune, tokens *[]tok.Token, errorReport *errorreport.ErrorReport) {
	// First consume the current token...
	r := runes[location.Current]
	// ...then move to the next token in preparation for the next loop
	// or for matching/peeking
	location.Current++

	switch r {
	case '(':
		addToken(tokens, tok.LeftParen, nil)
	case ')':
		addToken(tokens, tok.RightParen, nil)
	case '{':
		addToken(tokens, tok.LeftBrace, nil)
	case '}':
		addToken(tokens, tok.RightBrace, nil)
	case ',':
		addToken(tokens, tok.Comma, nil)
	case '.':
		addToken(tokens, tok.Dot, nil)
	case '-':
		addToken(tokens, tok.Minus, nil)
	case '+':
		addToken(tokens, tok.Plus, nil)
	case ';':
		addToken(tokens, tok.Semicolon, nil)
	case '*':
		addToken(tokens, tok.Star, nil)
	case '!':
		if match('=', location, runes) {
			addToken(tokens, tok.BangEqual, nil)
		} else {
			addToken(tokens, tok.Bang, nil)
		}
	case '=':
		if match('=', location, runes) {
			addToken(tokens, tok.EqualEqual, nil)
		} else {
			addToken(tokens, tok.Equal, nil)
		}
	case '<':
		if match('=', location, runes) {
			addToken(tokens, tok.LessEqual, nil)
		} else {
			addToken(tokens, tok.Less, nil)
		}
	case '>':
		if match('=', location, runes) {
			addToken(tokens, tok.GreaterEqual, nil)
		} else {
			addToken(tokens, tok.Greater, nil)
		}
	case '/':
		if match('/', location, runes) {
			// Keep advancing to end of comment line
			for peek(location, runes) != '\n' && !location.isAtEnd(runes) {
				location.Current++
			}
		} else {
			addToken(tokens, tok.Slash, nil)
		}
	case ' ', '\r', '\t':
		// Ignore whitespace
	case '\n':
		location.Line++
	case '"':
		handleString(location, tokens, runes, errorReport)
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		handleNumber(location, tokens, runes, errorReport)
	default:
		// At this point it's either an identifier or an unexpected
		// character.
		if isAlpha(r) {
			handleIdentifier(location, tokens, runes, errorReport)
		} else {
			errorReport.Report(location.Line, "", "Unexpected character.")
		}
	}
}

func handleIdentifier(location *sourceLocation, tokens *[]tok.Token, runes []rune, errorReport *errorreport.ErrorReport) {
	for isAlphaNumeric(peek(location, runes)) {
		location.Current++
	}

	tokenString := string(runes[location.Start:location.Current])
	tokenType, keyExists := keywords[tokenString]
	if keyExists {
		addToken(tokens, tokenType, nil)
	} else {
		addToken(tokens, tok.Identifier, nil)
	}
}

func handleString(location *sourceLocation, tokens *[]tok.Token, runes []rune, errorReport *errorreport.ErrorReport) {
	// Keep advancing to closing "
	for peek(location, runes) != '"' && !location.isAtEnd(runes) {
		if peek(location, runes) == '\n' {
			location.Line++
		}
		location.Current++
	}

	// Unterminated string
	if location.isAtEnd(runes) {
		errorReport.Report(location.Line, "", "Unterminated string.")
		return
	}

	// Consume the closing "
	location.Current++

	// Trim the surrounding quotes
	stringValue := string(runes[location.Start+1 : location.Current-1])
	addToken(tokens, tok.String, stringValue)
}

func handleNumber(location *sourceLocation, tokens *[]tok.Token, runes []rune, errorReport *errorreport.ErrorReport) {
	for isDigit(peek(location, runes)) {
		location.Current++
	}

	// Look for a fractional part
	if peek(location, runes) == '.' && isDigit(peekNext(location, runes)) {
		// Consume the "."
		location.Current++

		for isDigit(peek(location, runes)) {
			location.Current++
		}
	}

	numString := string(runes[location.Start:location.Current])
	numValue, err := strconv.ParseFloat(numString, 64)
	if err != nil {
		errorReport.Report(location.Line, "", "Could not convert number literal to float.")
		return
	}

	addToken(tokens, tok.Number, numValue)
}

func addToken(tokens *[]tok.Token, tokenType tok.TokenType, value any) {
	*tokens = append(*tokens, tok.Token{TokenType: tokenType, Literal: value})
}

func match(expected rune, location *sourceLocation, runes []rune) bool {
	if location.isAtEnd(runes) {
		return false
	}

	if runes[location.Current] != expected {
		return false
	}

	location.Current++
	return true
}

func peek(location *sourceLocation, runes []rune) rune {
	if location.isAtEnd(runes) {
		return 0
	}

	return runes[location.Current]
}

func peekNext(location *sourceLocation, runes []rune) rune {
	if location.Current+1 >= len(runes) {
		return 0
	}
	return runes[location.Current+1]
}

func isDigit(r rune) bool {
	return r >= 0x30 && r <= 0x39
}

func isAlpha(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_'
}

func isAlphaNumeric(r rune) bool {
	return isAlpha(r) || isDigit(r)
}
