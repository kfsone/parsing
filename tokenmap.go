package parsing

// TokenMap is used for the initial determination of what token or token-type a given
// ascii character [may] represent. For instance, byte(' ') is mapped here to
// WhitespaceToken, while the ascii numeric digits '0'-'9' are mapped to DigitToken, allowing
// the tokenizer to deduce which sub-tokenizer is appropriate based on the first
// character of a sequence.

// TokenMap maps individual ASCII character values to  base Token types.
var TokenMap = [256]Token{}

func init() {
	// Default to InvalidToken
	for i := 0; i < len(TokenMap); i++ {
		TokenMap[i] = InvalidToken
	}

	TokenMap[' '] = WhitespaceToken
	TokenMap['\t'] = WhitespaceToken
	TokenMap['\r'] = NewlineToken
	TokenMap['\n'] = NewlineToken
	for c := '0'; c <= '9'; c++ {
		TokenMap[c] = DigitToken
	}
	for c := 'a'; c <= 'z'; c++ {
		TokenMap[c] = AlphaToken
		TokenMap[c-'a'+'A'] = AlphaToken
	}
	TokenMap['{'] = OpenBrace
	TokenMap['}'] = CloseBrace
	TokenMap['('] = OpenParen
	TokenMap[')'] = CloseParen
	TokenMap['['] = OpenBracket
	TokenMap[']'] = CloseBracket
	TokenMap['.'] = Period
	TokenMap[','] = Comma
	TokenMap['$'] = Dollar
	TokenMap['+'] = Plus
	TokenMap['*'] = Asterisk
	TokenMap['-'] = Minus
	TokenMap[':'] = Colon
	TokenMap[';'] = Semicolon
	TokenMap['_'] = Underscore
	TokenMap['='] = Equals
	TokenMap['\''] = StringToken
	TokenMap['/'] = Slash
	TokenMap['"'] = StringToken

	for _, c := range []byte{'~', '!', '@', '#', '%', '^', '&', '\\', '|', '<', '>', '?'} {
		TokenMap[c] = SymbolToken
	}
}

// IsNumeric will return true if the given character is a digit or period.
func IsNumeric(char byte) bool {
	return TokenMap[char] == DigitToken || TokenMap[char] == Period
}

// IsIdentifierContinuation returns true for any tokens that are allowed to continue
// an identifier (2nd character+).
func IsIdentifierContinuation(char byte) bool {
	return TokenMap[char] == AlphaToken || TokenMap[char] == DigitToken || TokenMap[char] == Underscore
}

// IsAlpha returns true for any character a-z or A-Z
func IsAlpha(char byte) bool {
	return TokenMap[char] == AlphaToken
}
