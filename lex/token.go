package lex

type Token struct {
	kind  tokenKind
	loc   Location
	value string
}

func NewToken(kind tokenKind, loc Location, value string) *Token {
	return &Token{
		kind:  kind,
		loc:   loc,
		value: value,
	}
}

func (t *Token) equals(other *Token) bool {
	return t.value == other.value && t.kind == other.kind
}
