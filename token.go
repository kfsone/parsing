package parsing

// Token identifies a significant pattern in a code stream, from a specific
// keyword to an integer to whitespace to end-of-file. Here, Token is a
// pointer to a friendly name for the Token. Using a pointer allows for
// fast comparison etc operations while using a pointer allows for easy
// translation to human-friendly form.
type Token struct {
	*string
}

// IsTerminal will return true for tokens that describe a Terminal (single match).
func (t Token) IsTerminal() bool {
	return (*t.string)[0] >= 'a' && (*t.string)[0] <= 'z'
}

// NewToken will return a new Token with the friendly name given, with
// an all uppercase name.
func NewToken(label string) Token {
	if label[0] < 'A' || label[0] > 'Z' {
		panic(label + ": tokens must begin with a capital letter")
	}
	return Token{&label}
}

func (t Token) String() string {
	return *t.string
}

// Terminals represent an explicit character match and start with a lowercase character.

// NewTerminal will return a new Token with a friendly, lowercase name.
func NewTerminal(label string) Token {
	if label[0] < 'a' || label[0] > 'z' {
		panic(label + ": terminals must begin with a lowercase letter")
	}
	return Token{&label}
}

// SomethingToken denotes a class of token rather a match to a single
// explicit value. E.g. 'EOF' represents the absence of a character,
// 'WHITESPACE' is any number of space or tabs, etc.
var (
	// Noise
	InvalidToken    = NewToken("INVALID")
	EOFToken        = NewToken("EOF")
	WhitespaceToken = NewToken("WHITESPACE")
	NewlineToken    = NewToken("NEWLINE")
	CommentToken    = NewToken("COMMENT")

	// Intermediate classifications
	AlphaToken  = NewToken("ALPHA")
	DigitToken  = NewToken("DIGIT")
	SymbolToken = NewToken("SYMBOL")

	// Literals
	StringToken  = NewToken("STRING")
	IntegerToken = NewToken("INTEGER")
	FloatToken   = NewToken("FLOAT")

	// Identifiers
	IdentifierToken = NewToken("IDENTIFIER")
)

// Tokens resolving to a single literal value.
var (
	// Symbols
	OpenBrace    = NewTerminal("open-brace")
	CloseBrace   = NewTerminal("close-brace")
	OpenBracket  = NewTerminal("open-bracket")
	CloseBracket = NewTerminal("close-bracket")
	OpenParen    = NewTerminal("open-parens")
	CloseParen   = NewTerminal("close-parens")
	Asterisk     = NewTerminal("asterisk")
	Slash        = NewTerminal("slash")
	Period       = NewTerminal("period")
	Comma        = NewTerminal("comma")
	Dollar       = NewTerminal("dollar-sign")
	Plus         = NewTerminal("plus-sign")
	Minus        = NewTerminal("minus-sign")
	Colon        = NewTerminal("colon")
	Semicolon    = NewTerminal("semicolon")
	Underscore   = NewTerminal("underscore")
	Equals       = NewTerminal("equals-sign")
)

func IsSignificant(token Token) bool {
	return token != WhitespaceToken && token != NewlineToken && token != CommentToken
}
