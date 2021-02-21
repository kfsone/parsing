package parsing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSymbol_Equals(t *testing.T) {
	sym := &Symbol{Token: EOFToken}
	if assert.False(t, sym.Equals(Token{})) {
		assert.True(t, sym.Equals(EOFToken))
		assert.False(t, sym.Equals(NewToken("EOF")))
	}
}

func TestSymbol_String(t *testing.T) {
	s := &Symbol{}
	assert.Equal(t, "", s.String())
	s.Value = "a value"
	assert.Equal(t, "a value", s.String())
}

func TestSymbol_Identity(t *testing.T) {
	bang := NewTerminal("bang")
	tests := []struct {
		token Token
		value string
		want  string
	}{
		{InvalidToken, "asdf", `INVALID`},
		{EOFToken, "", `EOF`},
		{EOFToken, "sdfg", `EOF`},
		{WhitespaceToken, "dfgh", `WHITESPACE`},
		{NewlineToken, "\r\n\n\r", `NEWLINE`},
		{CommentToken, "/*xyz*/", `COMMENT`},

		{AlphaToken, "g", `ALPHA "g"`},
		{DigitToken, "3", `DIGIT "3"`},
		{SymbolToken, "!", `SYMBOL "!"`},

		{StringToken, "my value", `"my value"`},
		{IntegerToken, "42", `INTEGER "42"`},
		{FloatToken, "4.2", `FLOAT "4.2"`},

		{IdentifierToken, "hello world", `"hello world"`},

		// Terminal with and without a value
		{bang, "", `bang`},
		{bang, "xyz", `bang ("xyz")`},
	}

	for _, tt := range tests {
		t.Run(tt.token.String(), func(t *testing.T) {
			symbol := Symbol{Token: tt.token, Value: tt.value}
			assert.Equal(t, tt.want, symbol.Identity())
		})
	}
}
