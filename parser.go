package parsing

import (
	"fmt"
	"os"

	"github.com/kfsone/parsing/lib/stats"
)

type Rule struct {
	Sequence []Token
	Applies  Token
}

// Parser is a core implementation of a lexing-based parser implementation.
type Parser struct {
	Lexer   *Lexer
	current *Symbol
	ahead   []*Symbol
	rules   []Rule

	Tracing        bool
	VerboseTracing bool
}

// NewParser will construct a new parser instance and read-ahead the
// first two symbols.
func NewParser(l *Lexer, rules ...Rule) *Parser {
	// Create an ahead buffer with 2 nil entries for the first two reads.
	p := &Parser{
		Lexer:          l,
		current:        nil,
		ahead:          make([]*Symbol, 0, 64),
		rules:          rules,
		Tracing:        *stats.Verbose > 1,
		VerboseTracing: *stats.Verbose > 2,
	}
	p.readAhead()
	p.Next()
	return p
}

// Raise invokes the parser's error-handling function to report an error.
func (p *Parser) Raise(err error) {
	RaiseHandler(err)
}

// RaiseHandler is a default implementation of a Raise handler which prints the error to stderr..
func RaiseHandler(err error) {
	fmt.Fprintln(os.Stderr, err.Error())
	if stats.BumpCounter("errors", 1) > 16 {
		fmt.Fprintln(os.Stderr, "too many errors, terminating")
		os.Exit(22)
	}
}

// Current will return the current Symbol. At EOF, this will be a symbol
// with the EOFToken.
func (p *Parser) Current() *Symbol { return p.current }

// Peek will look ahead to the next lexed Symbol. At EOF, this will be a
// symbol with the EOFToken.
func (p *Parser) Peek() *Symbol { return p.ahead[0] }

// EOF returns true if we have reached end-of-file, that the current
// token is the EOFToken.
func (p *Parser) EOF() bool {
	return p.current.Token == EOFToken
}

// Locate returns a string describing the filename, line number and character
// of a symbol.
func (p *Parser) Locate(symbol *Symbol) string {
	return fmt.Sprintf("%s:%d:%d", p.Lexer.Filename(), p.Lexer.LineNo(symbol.StartOffset), p.Lexer.CharNo(symbol.StartOffset))
}

// Push injects Symbols into the Symbol at the current position,
// so that after this call p.Current will be symbols[0].
func (p *Parser) Push(symbols []*Symbol) {
	if len(symbols) == 0 {
		return
	}
	p.current, symbols = symbols[0], append(symbols[1:], p.current)
	p.ahead = append(symbols, p.ahead...)
}

// readAhead gets the next symbol from the lexer and adds it to the end of the
// 'ahead' buffer.
func (p *Parser) readAhead() {
	for {
		p.Lexer.Advance()
		if IsSignificant(p.Lexer.Token) {
			break
		}
	}
	p.ahead = append(p.ahead, &Symbol{p.Lexer.Token, p.Lexer.String(), p.Lexer.Start, p.Lexer.End})
}

func (r Rule) apply(p *Parser) bool {
	if p.current.Token != r.Sequence[0] {
		return false
	}
	for aheadNo, step := range r.Sequence[1:] { // note: aheadNo will be 0-based
		if aheadNo >= len(p.ahead) {
			p.readAhead()
		}
		if p.ahead[aheadNo].Token != step {
			return false
		}
	}

	p.current.Token = r.Applies
	// the first step is 'current', the rest are ahead, so len(sequence) -1.
	aheadCount := len(r.Sequence) - 1
	p.current.EndOffset = p.ahead[aheadCount-1].EndOffset
	// readjust the window
	p.current.Value = string(p.Lexer.code[p.current.StartOffset:p.current.EndOffset])
	p.readAhead()
	p.ahead = p.ahead[aheadCount:]

	return true
}

func (p *Parser) applyRules() {
	for _, rule := range p.rules {
		if rule.apply(p) {
			return
		}
	}
}

func (p *Parser) advance() Token {
	p.current, p.ahead = p.ahead[0], p.ahead[1:]
	p.readAhead()
	p.applyRules()
	return p.current.Token
}

// Next performs a read ahead and returns the new current token.
func (p *Parser) Next() (token Token) {
	for {
		token = p.advance()
		if IsSignificant(token) {
			return
		}
	}
}

// Expecting will return current symbol if it corresponds to one of the given
// tokens, or it will return a user-friendly error describing the expectation.
// Panics on an unexpected end-of-file.
func (p *Parser) Expecting(tokens ...Token) (*Symbol, error) {
	if len(tokens) == 0 {
		panic("must specify at least one token")
	}
	for _, token := range tokens {
		if p.current.Token == token {
			return p.current, nil
		}
	}

	// If this is because we reached EOF, make it fatal.
	if p.EOF() {
		panic(fmt.Errorf("%s: unexpected end-of-file", p.Locate(p.Current())))
	}

	// Current token is none of those expected, present the user a list of what
	// we thought they should provide.
	var expecting = tokens[0].String()
	if len(tokens) > 1 {
		expecting = "either " + expecting
		if len(tokens) > 2 {
			for _, token := range tokens[1 : len(tokens)-1] {
				expecting += ", " + token.String()
			}
			// oxford/serial comma
			expecting += ","
		}
		expecting += " or " + tokens[len(tokens)-1].String()
	}

	return p.current, p.SyntaxErrorf(p.current, "%s", expecting)
}

// OptionalSequence will attempt to match two or more tokens while allowing
// for non-significant tokens (whitespace, newlines, comments).
// In the case of a match, it will return the significant, matched tokens;
// In the case of a partial match, it will return a list of all the traversed
// tokens (suitable for calling Push() if you want to return to the original state),
// if there is no match, it will return nil, nil.
func (p *Parser) OptionalSequence(tokens ...Token) (seen []*Symbol, err error) {
	if len(tokens) < 2 {
		panic("invalid sequence length")
	}
	if !p.current.Equals(tokens[0]) {
		return nil, nil
	}
	seen = make([]*Symbol, 0, len(tokens)*2)
	matched := make([]*Symbol, 1, len(tokens))
	matched[0] = p.current
	for idx, want := range tokens[1:] {
		seen = append(seen, p.current)
		actual := p.Next()
		if actual != want {
			return seen, p.SyntaxErrorf(p.current, "%s after %s", want.String(), tokens[idx].String())
		}
		if IsSignificant(actual) {
			matched = append(matched, p.current)
		}
	}
	p.Next()
	return matched, nil
}

// Errorf formats an error based on a Symbol's location.
func (p *Parser) Errorf(symbol *Symbol, msg string, args ...interface{}) error {
	return fmt.Errorf("%s: %s: %s",
		p.Locate(symbol),
		fmt.Sprintf(msg, args...),
		symbol.Identity())
}

// SyntaxErrorf formats a syntax error based on a symbol.
func (p *Parser) SyntaxErrorf(symbol *Symbol, msg string, args ...interface{}) error {
	return p.Errorf(symbol, "syntax error: expected %s, got", fmt.Sprintf(msg, args...))
}

func (p *Parser) DuplicateErrorf(duplicate *Symbol, original *Symbol, originalParser *Parser, msg string, args ...interface{}) error {
	err := p.Errorf(duplicate, msg, args...)
	return fmt.Errorf("%w\n%s: \\-> previous occurrence of %q is here", err, originalParser.Locate(original), duplicate)
}

func (p *Parser) trace(what string, msg string, args ...interface{}) {
	if p.Tracing {
		if p.VerboseTracing {
			what = p.Locate(p.Current()) + " " + what
		}
		msg = fmt.Sprintf(msg, args...)
		fmt.Printf("%s  ( %s )  %s\n", what, p.Current().Identity(), msg)
	}
}

func (p *Parser) Trace(what string) {
	p.trace(fmt.Sprintf("%s@%d", what, p.Lexer.Start), "->  [ %s ]", p.Peek().Identity())
}

func (p *Parser) Note(what string, msg string, args ...interface{}) {
	p.trace(what, "note: "+msg, args...)
}
