package parsing

import (
	"bytes"
	"fmt"
)

// EOFByte will be returned by Read and Peek when EOF is reached, but
// may also be returned if the stream contains binary data.
var EOFByte = byte(0)

// Intercept is a callback that attempts to apply grammar-specific rules to a
// character's token lookup before we apply the baseline rules. Use l.Value
// to get the previous parts of the current token.
type Intercept func(l *Lexer) (processed bool)

// InterceptTable is a dictionary of lists of intercepts.
type InterceptTable map[Token][]Intercept

// Lexer tracks parsing of a source stream.
type Lexer struct {
	name       string
	code       []byte
	Start, End int // current Token bounds
	Token      Token
	keywords   map[string]Token
	intercepts InterceptTable
}

// Filename returns the name of the file this lexer is parsing.
func (l *Lexer) Filename() string { return l.name }

// Value returns the literal text of the current token.
func (l *Lexer) Value() []byte { return l.code[l.Start:l.End] }

// String returns a string copy of the literal value of the token.
func (l *Lexer) String() string { return string(l.Value()) }

// Position returns the start and end byte-offsets of the current token.
func (l *Lexer) Position() (int, int) { return l.Start, l.End }

// LineNo calculates the line-offset of a byte-offset within a source file.
func (l *Lexer) LineNo(pos int) int {
	return bytes.Count(l.code[:pos], []byte{'\n'}) + 1
}

// CharNo calculates the characters-from-start-of-line for a byte-offset within a source file.
func (l *Lexer) CharNo(pos int) int {
	lastCr := bytes.LastIndex(l.code[:pos], []byte{'\n'}) + 1
	return pos - lastCr + 1
}

// NewLexer constructs a new Lexer instance.
func NewLexer(name string, code []byte) *Lexer {
	return &Lexer{
		name:       name,
		code:       code,
		Start:      0,
		End:        0,
		Token:      InvalidToken,
		keywords:   nil,
		intercepts: nil,
	}
}

// AddKeyword registers a keyword Terminal with lexer so that the lexer will recognize it
// and return the specified token. Keyword identification is automatically performed on
// "words" - tokens that start with a letter or underscore.
func (l *Lexer) AddKeyword(keyword string, token Token) {
	if !token.IsTerminal() {
		panic(fmt.Sprintf("keywords must be represented by Terminals: %q is a Token", token.String()))
	}
	// Require a terminal rather than a token.
	if l.keywords == nil {
		l.keywords = make(map[string]Token)
	}
	l.keywords[keyword] = token
}

// Intercept associates a token-identifier function with a particular token type. During
// parsing, intercepts will be performed in the order they were registered.
func (l *Lexer) AddIntercept(token Token, intercept Intercept) {
	if l.intercepts == nil {
		l.intercepts = make(InterceptTable)
	}
	l.intercepts[token] = append(l.intercepts[token], intercept)
}

// Read will attempt to get the next byte from the source code. ok will be false when
// end-of-file is reached.
func (l *Lexer) Read() (byte, bool) {
	if l.End < len(l.code) {
		char := l.code[l.End]
		l.End++
		return char, true
	}
	return EOFByte, false
}

// Peek will attempt to look at the next character in the source code without advancing
// the read index. At EOF, ok will be false.
func (l *Lexer) Peek() (byte, bool) {
	if l.End < len(l.code) {
		return l.code[l.End], true
	}
	return EOFByte, false
}

// Skip moves the read pointer ahead one, it does not check for EOF.
func (l *Lexer) Skip() {
	l.End++
}

// Increment 'End' only if the next character (Peek()) matches.
func (l *Lexer) ForwardOn(char byte) bool {
	if next, ok := l.Peek(); ok && next == char {
		l.End++
		return true
	}
	return false
}

// Fatal reports a terminal parsing error at the current location in the file.
func (l *Lexer) Fatal(msg string, args ...interface{}) {
	prefix := fmt.Sprintf("%s:%d:%d", l.name, l.LineNo(l.Start), l.CharNo(l.Start))
	if l.End > l.Start+1 {
		prefix += fmt.Sprintf("-%d:%d", l.LineNo(l.End), l.CharNo(l.End))
	}
	panic(fmt.Sprintf(prefix+": error: "+msg, args...))
}

// SymbolizeComment will attempt to detect single- or multi-line comments and
// symbolize them.
func (l *Lexer) SymbolizeComment() {
	next, _ := l.Peek()
	switch next {
	case '/':
		l.Skip()
		l.Token = CommentToken
		l.SymbolizeSingleLineComment()

	case '*':
		l.Skip()
		l.Token = CommentToken
		l.SymbolizeMultiLineComment()
	}
}

// SymbolizeSingleLineComment will consume all characters until end of line.
func (l *Lexer) SymbolizeSingleLineComment() {
	for {
		if char, ok := l.Read(); !ok || char == '\n' {
			return
		}
	}
}

// SymbolizeMultiLineComment will read everything up the end of a multi-
// line comment, terminated by "*/". Invokes Fatal() if EOF is reached
// without seeing the end of a comment.
func (l *Lexer) SymbolizeMultiLineComment() {
	for {
		if char, ok := l.Read(); char != '*' {
			if !ok {
				l.Fatal("unterminated multiline comment")
			}
		} else if next, _ := l.Peek(); next == '/' {
			l.Skip()
			return
		}
	}
}

// SymbolizeString handles quoted strings by searching for the corresponding
// close quote character with allowing for escaped quotes (\', \").
// Does not validate escape sequences,
// Considers end-of-line before close-quote an error,
// Considers end-of-file before close-quote as fatal,
func (l *Lexer) SymbolizeString() {
loop:
	for {
		char, ok := l.Read()
		switch char {
		case EOFByte:
			if !ok {
				break loop
			}

		case '\'', '"':
			if char == l.code[l.Start] {
				return
			}

		case '\\':
			char, ok := l.Peek()
			if ok && char != '\n' && char != '\r' {
				l.Skip()
				continue
			}
			break loop

		case '\n', '\r':
			break loop
		}
	}
	l.Fatal("unterminated string/missing close-quote?")
}

// SymbolizeWord consumes a token starting with a letter or underscore and
// classifies using optional keyword lookups or as IdentifierToken or IdentifierToken.
func (l *Lexer) SymbolizeWord() {
	// Try and consume as an identifier, which requires we do not End on an alpha.
	for l.End < len(l.code) && IsIdentifierContinuation(l.code[l.End]) {
		l.Skip()
	}
	if l.End > l.Start {
		l.Token = IdentifierToken
		if l.keywords != nil {
			if keyword, ok := l.keywords[l.String()]; ok {
				l.Token = keyword
			}
		}
	}
}

// SymbolizeNumber attempts to classify numeric-looking tokens as either
// IntegerToken or FloatToken.
func (l *Lexer) SymbolizeNumber() {
	// allow <digits> [. [<digits>]]]
	for l.End < len(l.code) && TokenMap[l.code[l.End]] == DigitToken {
		l.Skip()
	}
	if l.End < len(l.code) && TokenMap[l.code[l.End]] == Period {
		l.Token = FloatToken
		l.Skip()
		for l.End < len(l.code) && TokenMap[l.code[l.End]] == DigitToken {
			l.Skip()
		}
	} else {
		l.Token = IntegerToken
	}
}

// attempt to apply intercept rules.
func (l *Lexer) intercept() bool {
	// Don't trust the user not to tamper properties of lexer.
	lCopy := *l
	if intercepts, ok := l.intercepts[l.Token]; ok {
		for _, intercept := range intercepts {
			if intercept(&lCopy) {
				// accept any changes.
				*l = lCopy
				return true
			}
		}
	}
	return false
}

// Advance will try to classify the next token in the stream.
func (l *Lexer) Advance() bool {
	l.Start = l.End
	if l.End >= len(l.code) {
		l.Token = EOFToken
		return false
	}

	// Move Start to the beginning of the new token, capture the character and move the
	// read token beyond it.
	char := l.code[l.Start]
	l.Skip()

	l.Token = TokenMap[char]
	if l.intercepts != nil && l.intercept() {
		return true
	}

	switch l.Token {
	case WhitespaceToken, NewlineToken:
		for l.End < len(l.code) && TokenMap[l.code[l.End]] == l.Token {
			l.Skip()
		}

	case AlphaToken, Underscore:
		l.SymbolizeWord()

	case Slash:
		l.SymbolizeComment()

	case StringToken:
		l.SymbolizeString()

	case DigitToken:
		l.SymbolizeNumber()

	case Plus, Minus:
		if l.End < len(l.code) && IsNumeric(l.code[l.End]) {
			l.SymbolizeNumber()
		}

	case Period:
		if l.End < len(l.code) && TokenMap[l.code[l.End]] == DigitToken {
			// Let symbolize number see the decimal
			l.End--
			l.SymbolizeNumber()
		}
	}

	return true
}
