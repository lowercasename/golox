package token

import (
	"testing"
)

func TestString(t *testing.T) {
	tok := Token{Type: STRING, Lexeme: "\"Hello, World!\"", Literal: "Hello, World!", Line: 40}
	if tok.String() != "STRING \"Hello, World!\" Hello, World!" {
		t.Fatalf("expected=STRING \"Hello, World!\" Hello, World!, got=%q", tok.String())
	}
	tok = Token{Type: NUMBER, Lexeme: "123.456", Literal: 123.456, Line: 40}
	if tok.String() != "NUMBER 123.456 123.456" {
		t.Fatalf("expected=NUMBER 123.456 123.456, got=%q", tok.String())
	}
	tok = Token{Type: IDENTIFIER, Lexeme: "foo", Literal: nil, Line: 40}
	if tok.String() != "IDENTIFIER foo <nil>" {
		t.Fatalf("expected=IDENTIFIER foo <nil>, got=%q", tok.String())
	}
}
