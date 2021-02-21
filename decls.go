// Decl is the AST base-type.
package parsing

// Decl is the base for any type of declaration, identifying where it came from.
type Decl struct {
	SourceFile string  `json:"src"`
	DeclType   *Symbol `json:"decltype"`
	Name       *Symbol `json:"name",omitempty`
}

func NewDecl(p *Parser, declType *Symbol) *Decl {
	return &Decl{SourceFile: p.Lexer.Filename(), DeclType: declType}
}
