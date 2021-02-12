package backend

import (
	"errors"

	"github.com/jameslahm/gosql/ast"
)

type ColumnType uint

const (
	TextType ColumnType = iota
	IntType
)

type Cell interface {
	AsText() string
	AsInt() int32
}

type ResultColumn struct {
	Name string
	Type ColumnType
}

type Results struct {
	Columns []ResultColumn
	Rows    [][]Cell
}

var (
	ErrTableDoesNotExist  = errors.New("Table does not exits")
	ErrColumnDoesNotExist = errors.New("Column does not exist")
	ErrInvalidSelectItem  = errors.New("Select item is not valid")
	ErrInvalidDataType    = errors.New("Invalid datatype")
	ErrMissingValues      = errors.New("Missing values")
)

type Backend interface {
	CreateTable(*ast.CreateTableStatement) error
	Insert(*ast.InsertStatement) error
	Select(*ast.SelectStatement) (*Results, error)
}
