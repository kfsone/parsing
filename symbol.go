package parsing

import (
	"encoding/json"
	"fmt"
)

// Symbol describes a specific instance of a token within the source stream.
type Symbol struct {
	// Token is how the lexer classified the symbol
	Token
	// Value is the literal string this represents.
	Value string
	// StartOffset is the byte-count to the first character of the Symbol.
	StartOffset int
	// EndOffset is the byte-count to the last character of the Symbol.
	EndOffset int
}

// Equals will test if a symbol represents a particular Token.
func (s *Symbol) Equals(t Token) bool { return s.Token == t }

// String returns the string representation of a Symbol's value.
func (s *Symbol) String() string { return s.Value }

// String will provide a string representation of a Symbol.
func (s *Symbol) Identity() string {
	if len(s.Value) == 0 {
		return *s.Token.string
	}
	if !s.Token.IsTerminal() {
		switch s.Token {
		case InvalidToken, EOFToken, WhitespaceToken, NewlineToken, CommentToken:
			return *s.Token.string
		case AlphaToken, DigitToken, SymbolToken, IntegerToken, FloatToken:
			return fmt.Sprintf("%s %q", *s.Token.string, s.String())
		case IdentifierToken, StringToken:
			return fmt.Sprintf("%q", s.String())
		}
	}
	return fmt.Sprintf("%s (%q)", *s.Token.string, s.String())
}

func (s *Symbol) MarshalJSON() (b []byte, e error) {
	token := s.Token.String()
	value := s.String()
	data := []string{token}
	if value != token {
		data = append(data, value)
	}
	typeName := "token"
	if s.Token.IsTerminal() {
		typeName = "terminal"
	}
	return json.Marshal(map[string][]string{typeName: data})
}
