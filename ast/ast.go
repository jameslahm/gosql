package ast

import "github.com/jameslahm/gosql/lex"

type Ast struct {
	Statements []*Statement
}

type Statement struct {
	SelectStatement      *SelectStatement
	CreateTableStatement *CreateTableStatement
	InsertStatement      *InsertStatement
	Kind                 AstKind
}

type SelectStatement struct {
	Items *[]*Expression
	From  lex.Token
}

type CreateTableStatement struct {
	Name lex.Token
	Cols *[]*ColumnDefinition
}

type ColumnDefinition struct {
	Name     lex.Token
	DataType lex.Token
}

type InsertStatement struct {
	table  lex.Token
	values *[]*Expression
}
