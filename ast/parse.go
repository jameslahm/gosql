package ast

import (
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
		for _, delimiter := range delimiters {
			if delimiter == token.Value {
				break
			}
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
				Kind:    literalKind,
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
