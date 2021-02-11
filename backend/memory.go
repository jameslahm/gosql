package backend

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"

	"github.com/jameslahm/gosql/ast"
	"github.com/jameslahm/gosql/lex"
)

type MemoryCell []byte

func (mc MemoryCell) AsInt() int {
	var i int
	err := binary.Read(bytes.NewBuffer(mc), binary.BigEndian, &i)
	if err != nil {
		panic(err)
	}
	return i
}

func (mc MemoryCell) AsText() string {
	return string(mc)
}

type Table struct {
	Columns     []string
	ColumnTypes []ColumnType
	Rows        [][]MemoryCell
}

type MemoryBackend struct {
	Tables map[string]*Table
}

func NewMemoryBackend() *MemoryBackend {
	return &MemoryBackend{
		Tables: make(map[string]*Table),
	}
}

func (mb *MemoryBackend) CreateTable(stmt *ast.CreateTableStatement) error {
	tables := mb.Tables
	var table Table
	tables[stmt.Name.Value] = &table
	for _, col := range *stmt.Cols {
		table.Columns = append(table.Columns, col.Name.Value)
		switch col.DataType.Value {
		case string(lex.IntKeyword):
			table.ColumnTypes = append(table.ColumnTypes, IntType)
		case string(lex.TextKeyword):
			table.ColumnTypes = append(table.ColumnTypes, TextType)
		default:
			return ErrInvalidDataType
		}
	}
	return nil
}

func (mb *MemoryBackend) Insert(stmt *ast.InsertStatement) error {
	table, ok := mb.Tables[stmt.Table.Value]
	if !ok {
		return ErrTableDoesNotExist
	}
	if len(table.Columns) != len(*stmt.Values) {
		return ErrMissingValues
	}

	if *stmt.Values == nil {
		return nil
	}

	var row []MemoryCell
	for _, value := range *stmt.Values {
		if value.Kind != ast.LiteralKind {
			fmt.Println("Skipping non-literal...")
			continue
		}
		row = append(row, mb.tokenToCell(value.Literal))
	}
	table.Rows = append(table.Rows, row)
	return nil
}

func (mb *MemoryBackend) tokenToCell(t *lex.Token) MemoryCell {
	if t.Kind == lex.NumberKind {
		value, err := strconv.Atoi(t.Value)
		if err != nil {
			panic(err)
		}
		var buf = new(bytes.Buffer)
		err = binary.Write(buf, binary.BigEndian, value)
		if err != nil {
			panic(err)
		}
		return buf.Bytes()
	} else if t.Kind == lex.StringKind {
		return []byte(t.Value)
	}
	return nil
}

func (mb *MemoryBackend) Select(stmt *ast.SelectStatement) (*Results, error) {
	table, ok := mb.Tables[stmt.From.Value]
	if !ok {
		return nil, ErrTableDoesNotExist
	}

	var columns []ResultColumn
	var indexes []int
	for _, col := range *stmt.Items {
		for i, colInfo := range table.Columns {
			if col.Literal.Value == colInfo {
				indexes = append(indexes, i)
				columns = append(columns, ResultColumn{
					Name: col.Literal.Value,
					Type: table.ColumnTypes[i],
				})
				break
			}
		}
	}

	if len(indexes) != len(*stmt.Items) {
		return nil, ErrColumnDoesNotExist
	}

	var resultRows [][]Cell
	for _, row := range table.Rows {
		var resultRow []Cell
		for _, i := range indexes {
			resultRow = append(resultRow, &row[i])
		}
		resultRows = append(resultRows, resultRow)
	}

	return &Results{
		Columns: columns,
		Rows:    resultRows,
	}, nil
}
