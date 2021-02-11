package lex

type Token struct {
	Kind  TokenKind
	Loc   Location
	Value string
}

func NewToken(kind TokenKind, loc Location, value string) *Token {
	return &Token{
		Kind:  kind,
		Loc:   loc,
		Value: value,
	}
}

func (t *Token) equals(other *Token) bool {
	return t.Value == other.Value && t.Kind == other.Kind
}
