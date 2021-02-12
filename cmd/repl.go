package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/jameslahm/gosql/ast"
	"github.com/jameslahm/gosql/backend"
	"github.com/jameslahm/gosql/lex"
)

func main() {
	mb := backend.NewMemoryBackend()
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Welcome to gosql")

	for {
		fmt.Print("$ ")
		text, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		text = strings.Replace(text, "\n", "", -1)
		tokens, err := lex.Lex(text)
		if err != nil {
			panic(err)
		}
		program, err := ast.Parse(tokens)
		if err != nil {
			panic(err)
		}
		for _, stmt := range program.Statements {
			switch stmt.Kind {
			case ast.CreateTableKind:
				err = mb.CreateTable(stmt.CreateTableStatement)
				if err != nil {
					panic(err)
				}
			case ast.InsertKind:
				err = mb.Insert(stmt.InsertStatement)
				if err != nil {
					panic(err)
				}
			case ast.SelectKind:
				results, err := mb.Select(stmt.SelectStatement)
				if err != nil {
					panic(err)
				}
				fmt.Printf("|")
				for _, col := range results.Columns {
					fmt.Printf("%10s|", col.Name)
				}
				fmt.Println()
				for _, row := range results.Rows {
					fmt.Printf("|")
					for i, cell := range row {
						switch results.Columns[i].Type {
						case backend.IntType:
							fmt.Printf("%10d|", cell.AsInt())
						case backend.TextType:
							fmt.Printf("%10s|", cell.AsText())
						}
					}
				}

			}
			fmt.Println("Ok")
		}
	}
}
