package parsing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_numbers(t *testing.T) {
	t.Run("0-9", func(t *testing.T) {
		matches := 0
		for i := '0'; i <= '9'; i++ {
			if TokenMap[i] == DigitToken {
				matches++
			}
		}
		assert.Equal(t, 10, matches)
	})
	t.Run("everything", func(t *testing.T) {
		misses := 0
		for i := 0; i < len(TokenMap); i++ {
			if TokenMap[i] != DigitToken {
				misses++
			}
		}
		assert.Equal(t, len(TokenMap)-10, misses)
	})
}

func Test_IsNumeric(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		assert.True(t, IsNumeric('0'))
		assert.True(t, IsNumeric('9'))
		assert.True(t, IsNumeric('.'))
	})
	t.Run("false", func(t *testing.T) {
		assert.False(t, IsNumeric('-'))
		assert.False(t, IsNumeric('+'))
		assert.False(t, IsNumeric(' '))
		assert.False(t, IsNumeric('a'))
	})
}

func Test_IsAlpha(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		assert.True(t, IsAlpha('a'))
		assert.True(t, IsAlpha('A'))
		assert.True(t, IsAlpha('g'))
		assert.True(t, IsAlpha('z'))
		assert.True(t, IsAlpha('Z'))
	})
	t.Run("false", func(t *testing.T) {
		assert.False(t, IsAlpha('-'))
		assert.False(t, IsAlpha('+'))
		assert.False(t, IsAlpha(' '))
		assert.False(t, IsAlpha('a'-1))
		assert.False(t, IsAlpha('z'+1))
		assert.False(t, IsAlpha('0'))
		assert.False(t, IsAlpha('9'))
		assert.False(t, IsAlpha('_'))
	})
}

func Test_IsIdentifierContinuation(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		assert.True(t, IsIdentifierContinuation('a'))
		assert.True(t, IsIdentifierContinuation('A'))
		assert.True(t, IsIdentifierContinuation('g'))
		assert.True(t, IsIdentifierContinuation('z'))
		assert.True(t, IsIdentifierContinuation('Z'))
		assert.True(t, IsIdentifierContinuation('0'))
		assert.True(t, IsIdentifierContinuation('9'))
		assert.True(t, IsIdentifierContinuation('_'))
	})
	t.Run("false", func(t *testing.T) {
		assert.False(t, IsIdentifierContinuation('-'))
		assert.False(t, IsIdentifierContinuation('+'))
		assert.False(t, IsIdentifierContinuation(' '))
		assert.False(t, IsIdentifierContinuation('a'-1))
		assert.False(t, IsIdentifierContinuation('z'+1))
	})
}
