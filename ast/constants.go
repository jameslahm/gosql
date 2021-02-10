package ast

type AstKind uint

const (
	SelectKind AstKind = iota
	CreateTableKind
	InsertKind
)

