package parsing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDecl(t *testing.T) {
	p := NewParser(NewLexer("decl.test", []byte("hot potato")))
	s := &Symbol{
		Token:       StringToken,
		Value:       "'fake'",
		StartOffset: 10,
		EndOffset:   20,
	}
	d := NewDecl(p, s)
	if assert.NotNil(t, d) {
		assert.Equal(t, d, &Decl{
			SourceFile: "decl.test",
			DeclType:   s,
			Name:       nil,
		})
	}
}
