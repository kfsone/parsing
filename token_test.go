package parsing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewToken(t *testing.T) {
	assert.Panics(t, func() { NewToken("abc") })
	assert.Equal(t, "X123abc", *NewToken("X123abc").string)
}

func TestToken_String(t *testing.T) {
	token := NewToken("ABC")
	assert.Equal(t, "ABC", token.String())
}

func TestNewTerminal(t *testing.T) {
	assert.Panics(t, func() { NewTerminal("ABC") })
	assert.Equal(t, "x123ABC", *NewTerminal("x123ABC").string)
}

func TestToken_IsTerminal(t *testing.T) {
	token, terminal := NewToken("Yadda"), NewTerminal("yadda")
	if assert.False(t, token.IsTerminal()) {
		assert.True(t, terminal.IsTerminal())
	}
}

func TestIsSignificant(t *testing.T) {
	tests := []struct {
		name  string
		token Token
		want  bool
	}{
		{"space", WhitespaceToken, false},
		{"newline", NewlineToken, false},
		{"comment", CommentToken, false},
		{"<new whitespace>", NewToken("WHITESPACE"), true},
		{"EOF", EOFToken, true},
		{"Invalid", InvalidToken, true},
		{"String", StringToken, true},
		{"Identifier", IdentifierToken, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsSignificant(tt.token))
		})
	}
}
