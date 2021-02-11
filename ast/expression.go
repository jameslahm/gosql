package ast

import "github.com/jameslahm/gosql/lex"

type Expression struct {
	Literal *lex.Token
	Kind    ExpressKind
}
