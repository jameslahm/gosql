package ast

import (
	"errors"
	"fmt"

	"github.com/jameslahm/gosql/lex"
)

// ? Helprs
func isKeyword(token *lex.Token, keyword lex.Keyword) bool {
	return token.Kind == lex.KeywordKind && token.Value == string(keyword)
}

func isSymbol(token *lex.Token, symbol lex.Symbol) bool {
	return token.Kind == lex.SymbolKind && token.Value == string(symbol)
}

func helpMessage(tokens []*lex.Token, cursor uint, msg string) {
	c := tokens[int(cursor)]
	fmt.Sprintf("[%d %d]: %s, Got %s\n", c.Loc.Line, c.Loc.Col, msg, c.Value)
}

func expectKeyword(tokens []*lex.Token, cursor uint, keyword lex.Keyword) bool {
	if uint(len(tokens)) <= cursor {
		return false
	} else {
		return isKeyword(tokens[cursor], keyword)
	}
}

func expectSymbol(tokens []*lex.Token, cursor uint, symbol lex.Symbol) bool {
	if uint(len(tokens)) <= cursor {
		return false
	} else {
		return isSymbol(tokens[cursor], symbol)
	}
}

func parseSelectStatement(tokens []*lex.Token, cursor uint, delimiter string) (*SelectStatement, uint, bool) {
	newCursor := cursor
	if !expectKeyword(tokens, newCursor, lex.SelectKeyword) {
		return nil, cursor, false
	}
	newCursor++
	slct := &SelectStatement{}
	var exps []*Expression
	var ok bool
	exps, newCursor, ok = parseExpressions(tokens, newCursor, []string{"from", delimiter})
	if !ok {
		return nil, cursor, false
	} else {
		slct.Items = &exps
		if !expectKeyword(tokens, newCursor, lex.FromKeyword) {
			helpMessage(tokens, cursor, "Expected from")
			return nil, cursor, false
		}
		slct.From = *tokens[newCursor]
		newCursor++
		return slct, newCursor, true
	}
}

func parseExpressions(tokens []*lex.Token, cursor uint, delimiters []string) ([]*Expression, uint, bool) {
	newCursor := cursor
	var exps []*Expression
	for {
		if newCursor >= uint((len(tokens))) {
			return nil, cursor, false
		}

		// ? Look for delimiters
		token := tokens[newCursor]

		var breakFlag = false
		for _, delimiter := range delimiters {
			if delimiter == token.Value {
				breakFlag = true
				break
			}
		}
		if breakFlag {
			break
		}

		// ? Look for comma
		if len(exps) > 0 {
			if !expectSymbol(tokens, newCursor, lex.CommaSymbol) {
				helpMessage(tokens, newCursor, "Expected comma")
				return nil, cursor, false
			}
			newCursor++
		}

		// ? Add Expressions
		var exp *Expression
		var ok bool
		exp, newCursor, ok = parseExpression(tokens, cursor)
		if !ok {
			helpMessage(tokens, cursor, "Expected expression")
			return nil, cursor, false
		}

		exps = append(exps, exp)
	}
	return exps, newCursor, true
}

func parseExpression(tokens []*lex.Token, cursor uint) (*Expression, uint, bool) {
	newCursor := cursor
	kinds := []lex.TokenKind{lex.NumberKind, lex.StringKind, lex.IdentifierKind}
	for _, kind := range kinds {
		if token, newCursor, ok := parseToken(tokens, newCursor, kind); ok {
			return &Expression{
				Literal: token,
				Kind:    LiteralKind,
			}, newCursor, true
		}
	}
	return nil, cursor, false
}

func parseToken(tokens []*lex.Token, cursor uint, kind lex.TokenKind) (*lex.Token, uint, bool) {
	if uint(len(tokens)) <= cursor {
		return nil, cursor, false
	} else if tokens[cursor].Kind == kind {
		return tokens[cursor], cursor + 1, true
	} else {
		return nil, cursor, false
	}
}

func parseInsertStatement(tokens []*lex.Token, cursor uint, delimiter string) (*InsertStatement, uint, bool) {
	newCursor := cursor

	if !expectKeyword(tokens, newCursor, lex.InsertKeyword) {
		return nil, cursor, false
	}
	newCursor++

	if !expectKeyword(tokens, newCursor, lex.IntoKeyword) {
		return nil, cursor, false
	}
	newCursor++

	var table *lex.Token
	var ok bool
	if table, newCursor, ok = parseToken(tokens, newCursor, lex.IdentifierKind); !ok {
		helpMessage(tokens, newCursor, "Expected table name")
		return nil, cursor, false
	}

	if !expectKeyword(tokens, newCursor, lex.ValuesKeyword) {
		helpMessage(tokens, newCursor, "Expected values")
		return nil, cursor, false
	}
	newCursor++

	if !expectSymbol(tokens, newCursor, lex.LeftParenSymbol) {
		helpMessage(tokens, newCursor, "Expected (")
		return nil, cursor, false
	}
	newCursor++

	var exps []*Expression
	exps, newCursor, ok = parseExpressions(tokens, newCursor, []string{")"})

	if !expectSymbol(tokens, newCursor, lex.RightParenSymbol) {
		helpMessage(tokens, newCursor, "Expected )")
		return nil, cursor, false
	}
	newCursor++

	return &InsertStatement{
		Values: &exps,
		Table:  *table,
	}, newCursor, true
}

func parseCreateStatement(tokens []*lex.Token, cursor uint, delimiters string) (*CreateTableStatement, uint, bool) {
	newCursor := cursor

	if !expectKeyword(tokens, newCursor, lex.CreateKeyword) {
		return nil, cursor, false
	}
	newCursor++

	if !expectKeyword(tokens, newCursor, lex.TableKeyword) {
		return nil, cursor, false
	}
	newCursor++

	var table *lex.Token
	var ok bool
	if table, newCursor, ok = parseToken(tokens, newCursor, lex.IdentifierKind); !ok {
		helpMessage(tokens, newCursor, "Expected table name")
		return nil, cursor, false
	}

	if !expectSymbol(tokens, newCursor, lex.LeftParenSymbol) {
		helpMessage(tokens, newCursor, "Expected (")
		return nil, cursor, false
	}
	newCursor++

	var cols []*ColumnDefinition
	cols, newCursor, ok = parseColumnDefinitions(tokens, newCursor, []string{")"})
	if !ok {
		return nil, cursor, false
	}

	if !expectSymbol(tokens, newCursor, lex.RightParenSymbol) {
		helpMessage(tokens, newCursor, "Expected )")
		return nil, cursor, false
	}
	newCursor++

	return &CreateTableStatement{
		Name: *table,
		Cols: &cols,
	}, newCursor, true
}

func parseColumnDefinitions(tokens []*lex.Token, cursor uint, delimiters []string) ([]*ColumnDefinition, uint, bool) {
	newCursor := cursor

	var cols []*ColumnDefinition

	for {
		if uint(len(tokens)) <= newCursor {
			return nil, cursor, false
		}

		current := tokens[newCursor]
		newCursor++

		var breakFlag = false
		for _, delimiter := range delimiters {
			if delimiter == current.Value {
				breakFlag = true
				break
			}
		}
		if breakFlag {
			break
		}

		if len(cols) > 0 {
			if !expectSymbol(tokens, newCursor, lex.CommaSymbol) {
				helpMessage(tokens, newCursor, "Expected comma")
			}
			newCursor++
		}

		var name *lex.Token
		var ok bool
		name, newCursor, ok = parseToken(tokens, newCursor, lex.IdentifierKind)
		if !ok {
			helpMessage(tokens, newCursor, "Expected col name")
			return nil, cursor, false
		}

		var dataType *lex.Token
		dataType, newCursor, ok = parseToken(tokens, newCursor, lex.KeywordKind)
		if !ok {
			helpMessage(tokens, newCursor, "Expected col data type")
			return nil, cursor, false
		}

		cols = append(cols, &ColumnDefinition{
			Name:     *name,
			DataType: *dataType,
		})
	}
	return cols, newCursor, true
}

func parseStatement(tokens []*lex.Token, cursor uint, delimiter string) (*Statement, uint, bool) {
	newCursor := cursor

	var slct *SelectStatement
	var ok bool
	slct, newCursor, ok = parseSelectStatement(tokens, newCursor, ";")
	if ok {
		return &Statement{
			Kind:            SelectKind,
			SelectStatement: slct,
		}, newCursor, true
	}

	var inst *InsertStatement
	inst, newCursor, ok = parseInsertStatement(tokens, newCursor, ";")
	if ok {
		return &Statement{
			Kind:            InsertKind,
			InsertStatement: inst,
		}, newCursor, true
	}

	var crst *CreateTableStatement
	crst, newCursor, ok = parseCreateStatement(tokens, newCursor, ";")
	if ok {
		return &Statement{
			CreateTableStatement: crst,
			Kind:                 CreateTableKind,
		}, newCursor, true
	}

	return nil, cursor, false
}

func Parse(tokens []*lex.Token) (*Ast, error) {
	var a = Ast{}
	var cursor uint = 0
	for cursor < uint(len(tokens)) {
		stmt, newCursor, ok := parseStatement(tokens, cursor, ";")
		if !ok {
			return nil, errors.New("Failed to parse, expected statement")
		}
		cursor = newCursor

		a.Statements = append(a.Statements, stmt)

		var atLeastOneSemicolon = false

		for expectSymbol(tokens, cursor, lex.SemiColonSymbol) {
			cursor++
			atLeastOneSemicolon = true
		}

		if !atLeastOneSemicolon {
			return nil, errors.New("Expected semicolon delimiter between statements")
		}
	}

	return &a, nil
}
