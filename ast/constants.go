package ast

type AstKind uint

const (
	SelectKind AstKind = iota
	CreateTableKind
	InsertKind
)

type ExpressKind uint

const (
	literalKind ExpressKind = iota
)
