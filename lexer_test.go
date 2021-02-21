package parsing

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLexer_Filename(t *testing.T) {
	l := &Lexer{}
	if assert.Equal(t, l.Filename(), "") {
		l.name = "c:\\bar\\foo.txt"
		assert.Equal(t, l.Filename(), "c:\\bar\\foo.txt")
	}
}

func TestLexer_Value(t *testing.T) {
	code := []byte("012345678")
	l := &Lexer{code: code, Start: 0, End: 2}
	if assert.Equal(t, code[0:2], l.Value()) {
		l.Start = 3
		l.End = 7
		assert.Equal(t, code[3:7], l.Value())
	}
}

func TestLexer_String(t *testing.T) {
	l := &Lexer{code: []byte("hippo"), Start: 0, End: 3}
	if assert.Equal(t, "hip", l.String()) {
		l.Start, l.End = 2, 4
		assert.Equal(t, "pp", l.String())
	}
}

func TestLexer_Position(t *testing.T) {
	l := &Lexer{Start: 3, End: 798}
	start, end := l.Position()
	if assert.Equal(t, 3, start) {
		assert.Equal(t, 798, end)
	}
}

func TestLexer_LineNo(t *testing.T) {
	tests := []struct {
		name string
		code string
		pos  int
		want int
	}{
		{"empty/0", "", 0, 1},
		{"aaa/0", "aaa", 0, 1},
		{"aaa/1", "aaa", 1, 1},
		{"aaa/2", "aaa", 2, 1},
		{"aaa/3", "aaa", 3, 1},
		{"a\\na/0", "a\na", 0, 1},
		{"a\\na/1", "a\na", 1, 1},
		{"a\\na/2", "a\na", 2, 2},
		{"a\\na/3", "a\na", 3, 2},
		{"\\n\\n\\n/0", "\n\n\n", 0, 1},
		{"\\n\\n\\n/1", "\n\n\n", 1, 2},
		{"\\n\\n\\n/2", "\n\n\n", 2, 3},
		{"\\n\\n\\n/3", "\n\n\n", 3, 4},
		{"\\naa\\na/2", "\naa\na", 3, 2},
		{"\\naa\\na/3", "\naa\na", 4, 3},
		{"\\n\\n\\n\\na/5", "\n\n\n\na", 5, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLexer("test", []byte(tt.code))
			if got := l.LineNo(tt.pos); got != tt.want {
				t.Errorf("Lexer.LineNo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLexer_CharNo(t *testing.T) {

	tests := []struct {
		name string
		code string
		pos  int
		want int
	}{
		{"empty/0", "", 0, 1},
		{"aaa\\nb/0", "aaa\nb", 0, 1},
		{"aaa\\nb/1", "aaa\nb", 1, 2},
		{"aaa\\nb/2", "aaa\nb", 2, 3},
		{"aaa\\nb/3", "aaa\nb", 3, 4},
		{"aaa\\nb/4", "aaa\nb", 4, 1},
		{"aaa\\nb/5", "aaa\nb", 5, 2},
		{"\\r\\nab/0", "\r\nab", 0, 1},
		{"\\r\\nab/1", "\r\nab", 1, 2},
		{"\\r\\nab/2", "\r\nab", 2, 1},
		{"\\r\\nab/3", "\r\nab", 3, 2},
		{"\\r\\nab/4", "\r\nab", 4, 3},
		{"a\\n\\nb\\nc/2", "a\n\nb\nc", 2, 1},
		{"a\\n\\nb\\nc/3", "a\n\nb\nc", 3, 1},
		{"a\\n\\nb\\nc/4", "a\n\nb\nc", 4, 2},
		{"a\\n\\nb\\nc/5", "a\n\nb\nc", 5, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLexer("test", []byte(tt.code))
			if got := l.CharNo(tt.pos); got != tt.want {
				t.Errorf("Lexer.CharNo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_NewLexer(t *testing.T) {
	lexer := NewLexer("file", []byte("hello"))
	assert.NotNil(t, lexer)
	assert.Equal(t, "file", lexer.name)
	assert.EqualValues(t, []byte{'h', 'e', 'l', 'l', 'o'}, lexer.code)
	assert.Equal(t, lexer.Start, 0)
	assert.Equal(t, lexer.End, 0)
	assert.Equal(t, InvalidToken, lexer.Token)
	assert.Nil(t, lexer.keywords)
	assert.Nil(t, lexer.intercepts)
}

func TestLexer_AddKeyword(t *testing.T) {
	t.Run("require terminal v token", func(t *testing.T) {
		lexer := &Lexer{}
		token := NewToken("BISCUITS")
		assert.Panics(t, func() { lexer.AddKeyword("yadda", token) })
		terminal := NewTerminal("cookie")
		lexer.AddKeyword("yadda", terminal)
	})
	t.Run("functionality", func(t *testing.T) {
		lexer := &Lexer{}
		kw := NewTerminal("key word")
		lexer.AddKeyword("keyword", kw)
		if assert.NotNil(t, lexer.keywords) {
			assert.Len(t, lexer.keywords, 1)
			if assert.Contains(t, lexer.keywords, "keyword") {
				assert.Equal(t, kw, lexer.keywords["keyword"])
			}
		}
	})
}

func TestLexer_AddIntercept(t *testing.T) {
	lexer := &Lexer{}
	called := false
	lexer.AddIntercept(EOFToken, func(*Lexer) bool { called = true; return true })
	if assert.NotNil(t, lexer.intercepts) {
		assert.Len(t, lexer.intercepts, 1)
		if assert.Contains(t, lexer.intercepts, EOFToken) {
			intercepts := lexer.intercepts[EOFToken]
			assert.NotNil(t, intercepts)
			if assert.Len(t, intercepts, 1) {
				intercept := intercepts[0]
				if assert.NotNil(t, intercept) {
					assert.True(t, intercept(lexer))
					assert.True(t, called)
				}
			}
		}
	}
}

func TestLexer_Read_populated(t *testing.T) {
	lexer := NewLexer("test", []byte{'a', 'b', 'c'})
	require.NotNil(t, lexer)
	t.Run("1st(a)", func(t *testing.T) {
		char, ok := lexer.Read()
		if assert.True(t, ok) {
			assert.Equal(t, byte('a'), char)
			assert.Equal(t, 0, lexer.Start)
			assert.Equal(t, 1, lexer.End)
		}
	})

	t.Run("2nd(b)", func(t *testing.T) {
		char, ok := lexer.Read()
		if assert.True(t, ok) {
			assert.Equal(t, byte('b'), char)
			assert.Equal(t, 0, lexer.Start)
			assert.Equal(t, 2, lexer.End)
		}
	})

	t.Run("3rd(c)", func(t *testing.T) {
		char, ok := lexer.Read()
		if assert.True(t, ok) {
			assert.Equal(t, byte('c'), char)
			assert.Equal(t, 0, lexer.Start)
			assert.Equal(t, 3, lexer.End)
		}
	})

	t.Run("4th(eof)", func(t *testing.T) {
		char, ok := lexer.Read()
		if assert.False(t, ok) {
			assert.Equal(t, EOFByte, char)
			assert.Equal(t, 0, lexer.Start)
			assert.Equal(t, 3, lexer.End)
		}
	})
}

func TestLexer_Read_one(t *testing.T) {
	lexer := NewLexer("test", []byte{'1'})
	require.NotNil(t, lexer)
	t.Run("1st(1)", func(t *testing.T) {
		char, ok := lexer.Read()
		if assert.True(t, ok) {
			assert.Equal(t, byte('1'), char)
			assert.Equal(t, 0, lexer.Start)
			assert.Equal(t, 1, lexer.End)
		}
	})

	t.Run("2nd(eof)", func(t *testing.T) {
		char, ok := lexer.Read()
		if assert.False(t, ok) {
			assert.Equal(t, EOFByte, char)
			assert.Equal(t, 0, lexer.Start)
			assert.Equal(t, 1, lexer.End)
		}
	})
}

func TestLexer_Read_empty(t *testing.T) {
	lexer := NewLexer("test", []byte{})
	require.NotNil(t, lexer)
	char, ok := lexer.Read()
	if assert.False(t, ok) {
		assert.Equal(t, EOFByte, char)
		assert.Equal(t, 0, lexer.Start)
		assert.Equal(t, 0, lexer.End)
	}
}

func TestLexer_Peek_empty(t *testing.T) {
	lexer := NewLexer("test", []byte{})
	require.NotNil(t, lexer)
	char, ok := lexer.Peek()
	if assert.False(t, ok) {
		assert.Equal(t, EOFByte, char)
		assert.Equal(t, 0, lexer.Start)
		assert.Equal(t, 0, lexer.End)
	}
}

func TestLexer_Peek_populated(t *testing.T) {
	lexer := NewLexer("test", []byte{})
	require.NotNil(t, lexer)
	char, ok := lexer.Peek()
	if assert.False(t, ok) {
		assert.Equal(t, EOFByte, char)
		assert.Equal(t, 0, lexer.Start)
		assert.Equal(t, 0, lexer.End)
	}
}

func TestLexer_Skip(t *testing.T) {
	lexer := &Lexer{Start: 0, End: 1}
	lexer.Skip()
	if assert.Equal(t, lexer.End, 2) {
		lexer.Skip()
		assert.Equal(t, lexer.End, 3)
	}
}

func TestLexer_ForwardOn(t *testing.T) {
	lexer := &Lexer{code: []byte("hello"), Start: 1, End: 1}
	if assert.False(t, lexer.ForwardOn('h')) {
		assert.True(t, lexer.ForwardOn('e'))
		assert.False(t, lexer.ForwardOn('h'))
		assert.False(t, lexer.ForwardOn('e'))
		assert.True(t, lexer.ForwardOn('l'))
	}
}

func TestLexer_Fatal(t *testing.T) {
	t.Run("range", func(t *testing.T) {
		l := &Lexer{name: "mytest.txt", code: []byte("\nhello"), Start: 1, End: 6}
		assert.PanicsWithValue(t, "mytest.txt:2:1-2:6: error: goes boom", func() { l.Fatal("goes %s", "boom") })
	})

	t.Run("0-point", func(t *testing.T) {
		l := &Lexer{name: "aaa", code: []byte("\nhello"), Start: 0, End: 0}
		assert.PanicsWithValue(t, "aaa:1:1: error: badda-boom", func() { l.Fatal("%s-boom", "badda") })
	})

	t.Run("1-point", func(t *testing.T) {
		l := &Lexer{name: "stupid:name:for:a:file:", code: []byte("\nhello"), Start: 4, End: 5}
		assert.PanicsWithValue(t, "stupid:name:for:a:file::2:4: error: multipass", func() { l.Fatal("multipass") })
	})
}

func TestLexer_SymbolizeComment(t *testing.T) {
	tests := []struct {
		name, code   string
		offset, want int
	}{
		{"from 0 with nl", "//012345\r\n89", 0, 10},
		{"from 1 with nl", "//012345\r\n89", 1, 10},
		{"nl at start", "/\n\n\n\n\n", 0, 2},
		{"with comments", "/// /**/ //", 0, 11},

		{"empty", "**/", 0, 3},
		{"two line", "*01\r\n45*/", 0, 9},
		{"multi-line with comments", "*/*\n//\n/*\n*/", 0, 12},
		{"offset", " /*01/*45*/", 2, 11},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := &Lexer{code: []byte(tt.code), Start: 0, End: tt.offset}
			lexer.SymbolizeComment()
			assert.Equal(t, tt.want, lexer.End)
		})
	}
}

func TestLexer_SymbolizeMultiLineComment(t *testing.T) {
	t.Run("panic unterminated", func(t *testing.T) {
		t.Run("populated", func(t *testing.T) {
			lexer := NewLexer("unterminated", []byte("abc"))
			if assert.NotNil(t, lexer) {
				assert.Panics(t, func() { lexer.SymbolizeMultiLineComment() })
			}
		})

		t.Run("unpopulated", func(t *testing.T) {
			lexer := NewLexer("unterminated", []byte(""))
			if assert.NotNil(t, lexer) {
				assert.Panics(t, func() { lexer.SymbolizeMultiLineComment() })
			}
		})
	})
}

func TestLexer_SymbolizeString(t *testing.T) {
	quotes := []byte{'\'', '"'}
	for idx, quote := range quotes {
		t.Run(string(quote), func(t *testing.T) {
			nonQuote := quotes[1-idx]
			t.Run("fail", func(t *testing.T) {
				tests := []string{
					"q", "q ", "q\n", "q\r", "q\r\n", "q\\", "qQ", "q Q", "q\\Q", "q\\q",
				}
				for _, str := range tests {
					str = strings.ReplaceAll(str, "q", string(quote))
					str = strings.ReplaceAll(str, "Q", string(nonQuote))
					t.Run(str, func(t *testing.T) {
						l := NewLexer("string.test", []byte(str))
						l.Skip()
						assert.Panics(t, func() { l.SymbolizeString() })
					})
				}
			})
			t.Run("pass", func(t *testing.T) {
				tests := []string{
					"qq", "qQq", "q\\qq", "qQ\\Q\\q\\\\q",
					"q/*\\q*/q", "qHello\\, \\QWorld\\Q!q",
				}
				for _, str := range tests {
					str = strings.ReplaceAll(str, "q", string(quote))
					str = strings.ReplaceAll(str, "Q", string(nonQuote))
					code := []byte(str + " garbage")
					t.Run(str, func(t *testing.T) {
						l := NewLexer("string.test", code)
						l.Skip()
						l.SymbolizeString()
						assert.Equal(t, code[0:len(str)], l.Value())
					})
				}
			})
		})
	}
}

func TestLexer_SymbolizeWord(t *testing.T) {
	// keyword token for testing that keywords work.
	asif := NewTerminal("asif")
	tests := []struct {
		code  string
		want  string
		token Token
	}{
		{"a!", "a", IdentifierToken},
		{"zb!", "zb", IdentifierToken},
		{"xy1_2:q", "xy1_2", IdentifierToken},
		{"asif.2", "asif", asif},
	}
	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			l := NewLexer("word.test", []byte(tt.code))
			l.AddKeyword("asif", asif)
			// "read" the first letter
			l.Skip()
			l.Token = AlphaToken
			l.SymbolizeWord()
			if assert.Equal(t, tt.token, l.Token) {
				assert.Equal(t, tt.want, l.String())
			}
		})
	}
}

func TestLexer_SymbolizeNumber(t *testing.T) {
	tests := []struct {
		code, want string
		token      Token
	}{
		{"1", "1", IntegerToken},
		{".3", ".3", FloatToken},
		{"5.", "5.", FloatToken},
		{"135", "135", IntegerToken},
		{"+123", "+123", IntegerToken},
		{"-.35", "-.35", FloatToken},
		{"0000.11111.234", "0000.11111", FloatToken},
	}
	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			l := NewLexer("number.test", []byte(tt.code))
			// "read" the first letter
			if tt.code[0] != '.' {
				// period gets shown thru
				l.Skip()
				l.Token = DigitToken
			} else {
				l.Token = Period
			}
			l.SymbolizeNumber()
			if assert.Equal(t, tt.token, l.Token) {
				assert.Equal(t, tt.want, l.String())
			}
		})
	}
}

func TestLexer_intercept(t *testing.T) {
	intercepts := make(InterceptTable)
	intercepts[Plus] = []Intercept{
		func(l *Lexer) bool {
			l.Start = 1001
			l.End = 1002
			return false
		},
	}
	intercepts[Minus] = []Intercept{
		func(l *Lexer) bool {
			l.Start = 1111
			l.End = 2222
			return false
		},
		func(l *Lexer) bool {
			l.Start = 2001
			l.End = 2002
			l.Token = IntegerToken
			return true
		},
	}
	testCases := []struct {
		name      string
		token     Token
		want      bool
		start     int
		end       int
		wantToken Token
	}{
		{"none", CommentToken, false, 0, 0, CommentToken},
		{"false", Plus, false, 0, 0, Plus},
		{"positive", Minus, true, 2001, 2002, IntegerToken},
	}
	for _, tc := range testCases {
		l := &Lexer{
			Start:      0,
			End:        0,
			Token:      tc.token,
			intercepts: intercepts,
		}
		if assert.Equal(t, tc.want, l.intercept()) {
			assert.Equal(t, l.Start, tc.start)
			assert.Equal(t, l.End, tc.end)
			assert.Equal(t, l.Token, tc.wantToken)
		}
	}
}

func TestLexer_Advance(t *testing.T) {
	tests := []struct {
		name             string
		code             string
		Start, End       int
		want             bool
		token            Token
		newStart, newEnd int
	}{
		{"at zero eof", "", 0, 0, false, EOFToken, 0, 0},
		{"past zero EOF", "", 1, 1, false, EOFToken, 1, 1},
		{"start one whitespace", " ", 0, 0, true, WhitespaceToken, 0, 1},
		{"start three newlines", "\n\r\n", 0, 0, true, WhitespaceToken, 0, 2},
		{"plus", "+", 0, 0, true, Plus, 0, 1},
		{"minus", "-", 0, 0, true, Minus, 0, 1},
		{"plus1", "+1a", 0, 0, true, IntegerToken, 0, 2},
		{"minus.0", "-.0a", 0, 0, true, FloatToken, 0, 3},
		{".a", ".a", 0, 0, true, Period, 0, 1},
		{".999a", ".999a", 0, 0, true, FloatToken, 0, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Lexer{
				name:  "advance.test",
				code:  []byte(tt.code),
				Start: tt.Start,
				End:   tt.End,
			}
			if got := l.Advance(); got != tt.want {
				t.Errorf("Advance() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("intercepts", func(t *testing.T) {
		l := NewLexer("intercept.test", []byte(":+-"))
		calls := 0
		l.AddIntercept(Plus, func(_ *Lexer) bool {
			calls++
			return false
		})
		l.AddIntercept(Minus, func(_ *Lexer) bool {
			calls++
			return true
		})
		t.Run("no match", func(t *testing.T) {
			if assert.True(t, l.Advance()) {
				assert.Equal(t, 0, calls)
			}
		})
		t.Run("returns false", func(t *testing.T) {
			if assert.True(t, l.Advance()) {
				assert.Equal(t, 1, calls)
			}
		})
		t.Run("returns true", func(t *testing.T) {
			if assert.True(t, l.Advance()) {
				assert.Equal(t, 2, calls)
			}
		})
	})
}
