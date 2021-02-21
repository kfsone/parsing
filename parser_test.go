package parsing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParser_Current(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		p := &Parser{}
		assert.Nil(t, p.Current())
	})

	t.Run("populated", func(t *testing.T) {
		symbol := &Symbol{}
		p := &Parser{current: symbol}
		assert.Equal(t, symbol, p.Current())
	})
}

func TestParser_Peek(t *testing.T) {
	symbol1, symbol2, symbol3 := &Symbol{}, &Symbol{}, &Symbol{}
	p := &Parser{current: symbol1, ahead: []*Symbol{symbol2, symbol3}}
	assert.Equal(t, symbol2, p.Peek())
}

func TestParser_EOF(t *testing.T) {
	p := &Parser{current: &Symbol{Token: NewlineToken}}
	if assert.False(t, p.EOF()) {
		p.current.Token = EOFToken
		assert.True(t, p.EOF())
	}
}

func TestParser_Locate(t *testing.T) {
	code := "01234\n67\n9"
	lexer := NewLexer("locate.test", []byte(code))
	parser := &Parser{Lexer: lexer}
	symbol := &Symbol{StartOffset: 6, EndOffset: 7}
	location := parser.Locate(symbol)
	assert.Equal(t, "locate.test:2:1", location)
}

func TestParser_Push(t *testing.T) {
	symbol1, symbol2, symbol3, symbol4 := &Symbol{}, &Symbol{}, &Symbol{}, &Symbol{}
	parser := &Parser{current: symbol4}
	t.Run("empty/noop", func(t *testing.T) {
		parser.Push([]*Symbol{})
		if assert.Equal(t, symbol4, parser.current) {
			assert.Nil(t, parser.ahead)
		}
	})
	t.Run("symbol3", func(t *testing.T) {
		parser.Push([]*Symbol{symbol3})
		if assert.Equal(t, symbol3, parser.current) {
			assert.Equal(t, []*Symbol{symbol4}, parser.ahead)
		}
	})
	// Confirm no deduplication.
	t.Run("symbol3 repeat", func(t *testing.T) {
		parser.Push([]*Symbol{symbol3})
		if assert.Equal(t, symbol3, parser.current) {
			assert.Equal(t, []*Symbol{symbol3, symbol4}, parser.ahead)
		}
	})
	t.Run("symbol1+2", func(t *testing.T) {
		parser.Push([]*Symbol{symbol1, symbol2})
		if assert.Equal(t, symbol1, parser.current) {
			assert.Equal(t, []*Symbol{symbol2, symbol3, symbol3, symbol4}, parser.ahead)
		}
	})
}

func Test_NewParser(t *testing.T) {
	code := []byte("//\n.'hello'")
	lexer := NewLexer("newparser.test", code)
	p := NewParser(lexer)
	require.NotNil(t, p)

	// Should have skipped the comment/whitespace.
	expectCurrent := Symbol{Token: Period, Value: ".", StartOffset: 3, EndOffset: 4}
	expectAhead := Symbol{Token: StringToken, Value: "'hello'", StartOffset: 4, EndOffset: 11}

	expect := Parser{
		Lexer:          lexer,
		current:        &expectCurrent,
		ahead:          []*Symbol{&expectAhead},
		rules:          nil,
		Tracing:        false,
		VerboseTracing: false,
	}
	assert.EqualValues(t, expect, *p)
}

func Test_Parser_readAhead(t *testing.T) {
	code := []byte("//\n'hello' 123\t()")
	lexer := NewLexer("newparser.test", code)
	p := &Parser{Lexer: lexer}
	t.Run("first token", func(t *testing.T) {
		p.readAhead()
		if assert.NotNil(t, p.ahead) {
			// should not have been touched.
			assert.Nil(t, p.current)
			assert.Len(t, p.ahead, 1)
			assert.Equal(t, 3, p.ahead[0].StartOffset)
			assert.Equal(t, 10, p.ahead[0].EndOffset)
			if assert.True(t, p.ahead[0].Equals(StringToken)) {
				// "'hello'"
				assert.Equal(t, string(code[3:10]), p.ahead[0].Value)
			}
		}
	})
	t.Run("second token", func(t *testing.T) {
		p.readAhead()
		if assert.NotNil(t, p.ahead) {
			// should not have been touched.
			assert.Nil(t, p.current)
			assert.Len(t, p.ahead, 2)
			// first token should still be in-place
			if assert.True(t, p.ahead[0].Equals(StringToken)) {
				assert.Equal(t, string(code[3:10]), p.ahead[0].Value)
			}
			// second token should be the number 123
			assert.Equal(t, 11, p.ahead[1].StartOffset)
			assert.Equal(t, 14, p.ahead[1].EndOffset)
			if assert.True(t, p.ahead[1].Equals(IntegerToken)) {
				assert.Equal(t, string(code[11:14]), p.ahead[1].Value)
			}
		}
	})
	t.Run("to EOF", func(t *testing.T) {
		p.readAhead()
		p.readAhead()
		p.readAhead()
		p.readAhead()
		if assert.NotNil(t, p.ahead) {
			assert.Len(t, p.ahead, 6)
			assert.True(t, p.ahead[2].Equals(OpenParen))
			assert.True(t, p.ahead[3].Equals(CloseParen))
			assert.True(t, p.ahead[4].Equals(EOFToken))
			assert.True(t, p.ahead[5].Equals(EOFToken))
		}
	})
}

func TestParser_Next(t *testing.T) {
	t.Run("skip leading space", func(t *testing.T) {
		p := NewParser(NewLexer("next.test", []byte("\t.")))
		assert.True(t, p.current.Equals(Period))
	})
	t.Run("skip leading newline", func(t *testing.T) {
		p := NewParser(NewLexer("next.test", []byte("\r\n.")))
		assert.True(t, p.current.Equals(Period))
	})
	t.Run("skip leading comment + whitespace", func(t *testing.T) {
		p := NewParser(NewLexer("next.test", []byte("// one line\n/*\nmulti\n*/\t \t:")))
		assert.True(t, p.current.Equals(Colon))
	})

	lexer := NewLexer("next.test", []byte("//comment\n+\t123 'do'"))
	p := NewParser(lexer)
	require.NotNil(t, p)
	require.True(t, p.current.Equals(Plus))
	require.True(t, p.ahead[0].Equals(IntegerToken))

	t.Run("first", func(t *testing.T) {
		token := p.Next()
		if assert.Equal(t, IntegerToken, token) {
			assert.Equal(t, "123", p.current.String())
		}
		assert.True(t, p.ahead[0].Equals(StringToken))
	})

	t.Run("second", func(t *testing.T) {
		token := p.Next()
		if assert.Equal(t, StringToken, token) {
			assert.Equal(t, "'do'", p.current.String())
		}
		assert.True(t, p.ahead[0].Equals(EOFToken))
	})

	t.Run("third", func(t *testing.T) {
		token := p.Next()
		assert.Equal(t, EOFToken, token)
		assert.True(t, p.ahead[0].Equals(EOFToken))
	})
}

func TestParser_Expecting(t *testing.T) {
	t.Run("catch zero tokens", func(t *testing.T) {
		p := &Parser{}
		assert.Panics(t, func() { p.Expecting() })
	})
	t.Run("EOF on no significant tokens", func(t *testing.T) {
		p := NewParser(NewLexer("expect.text", []byte("  // comment\r\n/*\nmulti\rline\ncomment*/\t\t\n")))
		if assert.True(t, p.Current().Equals(EOFToken)) {
			// If we ask explicitly for a whitespace token, it should be seen.
			_, err := p.Expecting(EOFToken)
			if assert.Nil(t, err) {
				assert.Panics(t, func() { p.Expecting(IdentifierToken) })
			}
		}
	})

	tests := []struct {
		name, code string
		tokens     []Token
		want       bool
		token      Token
		err        string
	}{
		{"match single", "123\n", []Token{IntegerToken}, true, IntegerToken, ""},
		{"match first", "[!", []Token{OpenBracket, SymbolToken}, true, OpenBracket, ""},
		{"match nth", "[!", []Token{CommentToken, SymbolToken, NewlineToken, EOFToken, OpenBracket}, true, OpenBracket, ""},
		{"non-match single", "hi", []Token{CommentToken}, false, Token{}, "expect.test:1:1: syntax error: expected COMMENT, got: \"hi\""},
		{"non-match dual", "hi", []Token{CommentToken, NewlineToken}, false, Token{}, "expect.test:1:1: syntax error: expected either COMMENT or NEWLINE, got: \"hi\""},
		{"non-match multi", "hi", []Token{CommentToken, WhitespaceToken, NewlineToken}, false, Token{}, "expect.test:1:1: syntax error: expected either COMMENT, WHITESPACE, or NEWLINE, got: \"hi\""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer("expect.test", []byte(tt.code))
			p := NewParser(lexer)
			symbol, err := p.Expecting(tt.tokens...)
			if tt.want {
				if assert.NotNil(t, symbol) && symbol != nil {
					assert.Equal(t, tt.token, symbol.Token)
					assert.Nil(t, err)
				}
			} else {
				assert.NotNil(t, err)
				assert.Equal(t, tt.err, err.Error())
			}
		})
	}
}

func TestParser_Errorf(t *testing.T) {
	code := "01234\n-> symbol\n9"
	lexer := NewLexer("tests/errorf.test", []byte(code))
	parser := &Parser{Lexer: lexer}
	symbol := &Symbol{Token: IdentifierToken, StartOffset: 9, EndOffset: 15, Value: "symbol"}

	err := parser.Errorf(symbol, "error %s", "msg")
	if assert.NotNil(t, err) {
		line1 := "tests/errorf.test:2:4: error msg: \"symbol\""
		assert.Equal(t, line1, err.Error())
	}
}

func TestParser_SyntaxErrorf(t *testing.T) {
	code := "01234\n-> symbol\n9"
	lexer := NewLexer("tests/syntaxerrorf.test", []byte(code))
	parser := &Parser{Lexer: lexer}
	symbol := &Symbol{Token: IdentifierToken, StartOffset: 9, EndOffset: 15, Value: "symbol"}

	err := parser.SyntaxErrorf(symbol, "error %s", "msg")
	if assert.NotNil(t, err) {
		line1 := "tests/syntaxerrorf.test:2:4: syntax error: expected error msg, got: \"symbol\""
		assert.Equal(t, line1, err.Error())
	}
}

func TestParser_DuplicateErrorf(t *testing.T) {
	t.Run("same parser", func(t *testing.T) {
		code := "// duplicate keyword\nmonkey see\nmonkey do\n"
		p := NewParser(NewLexer("tests/duplicateerrorf.test", []byte(code)))
		first := p.Current()
		if assert.Equal(t, "monkey", first.String()) {
			p.Next()
			p.Next()
			second := p.Current()
			if assert.Equal(t, "monkey", second.String()) {
				err := p.DuplicateErrorf(second, first, p, "repetition of %s", "noun")

				if assert.NotNil(t, err) {
					line1 := "tests/duplicateerrorf.test:3:1: repetition of noun: \"monkey\""
					line2 := "tests/duplicateerrorf.test:2:1: \\-> previous occurrence of \"monkey\" is here"
					assert.Equal(t, line1+"\n"+line2, err.Error())
				}
			}
		}
	})

	t.Run("different parser", func(t *testing.T) {
		p1 := NewParser(NewLexer("tests/duplicateerrorf.test", []byte("monkey see")))
		p2 := NewParser(NewLexer("differentfile.test", []byte("// skip me\n// and me\n\t\t monkey do")))
		first := p1.Current()
		second := p2.Current()
		if assert.Equal(t, "monkey", first.String()) {
			if assert.Equal(t, "monkey", second.String()) {
				err := p2.DuplicateErrorf(second, first, p1, "another %s", "noun")

				if assert.NotNil(t, err) {
					line1 := "differentfile.test:3:4: another noun: \"monkey\""
					line2 := "tests/duplicateerrorf.test:1:1: \\-> previous occurrence of \"monkey\" is here"
					assert.Equal(t, line1+"\n"+line2, err.Error())
				}
			}
		}
	})
}

func TestParser_Raise(t *testing.T) {

}
