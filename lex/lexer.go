package lex

import (
	"bytes"
	"fmt"
)

type lexer func(string, Cursor) (*Token, Cursor, bool)

func lex(source string) ([]*Token, error) {
	cursor := NewCursor(0, NewLocation())
	lexers := []lexer{lexKeyword, lexSymbol, lexString, lexNumber, lexIdentifier}
	tokens := []*Token{}
	for cursor.pointer < uint(len(source)) {
		for _, lexer := range lexers {
			if token, newCursor, ok := lexer(source, cursor); ok {
				cursor = newCursor
				if token != nil {
					tokens = append(tokens, token)
				}
				break
			}
		}
		hint := ""
		if len(tokens) > 0 {
			hint = fmt.Sprintf(" after %s", tokens[len(tokens)-1].value)
		}
		return nil, fmt.Errorf("Unable to lex token%s, at %d %d", hint, cursor.loc.line, cursor.loc.col)
	}
	return tokens, nil
}

func lexNumber(source string, cursor Cursor) (*Token, Cursor, bool) {
	newCursor := cursor
	periodFound := false
	expMarkerFound := false

	for ; newCursor.pointer < uint(len(source)); newCursor.pointer++ {
		character := source[newCursor.pointer]
		newCursor.loc.col++

		isDigit := character >= '0' && character <= '9'
		isPeriod := character == '.'
		isExpMarker := character == 'e'

		// ? Must start with period or digit
		if cursor.pointer == newCursor.pointer {
			if !isDigit && !isPeriod {
				return nil, cursor, false
			}
			if isPeriod {
				periodFound = true
			}
			continue
		}

		if isPeriod {
			if periodFound {
				return nil, cursor, false
			}
			periodFound = true
			continue
		}

		if isExpMarker {
			if expMarkerFound {
				return nil, cursor, false
			}
			expMarkerFound = true

			// ? expMarker must followed by number or +-
			if newCursor.pointer == uint(len(source)-1) {
				return nil, cursor, false
			}

			characterNext := source[newCursor.pointer+1]
			if characterNext == '+' || characterNext == '-' {
				newCursor.pointer++
				newCursor.loc.col++
			}
			continue
		}

		// ? Space
		if !isDigit {
			break
		}
	}

	if newCursor.pointer == cursor.pointer {
		return nil, cursor, false
	}
	return NewToken(numberKind, cursor.loc, source[cursor.pointer:newCursor.pointer]), newCursor, true
}

func lexCharacterDelimited(source string, cursor Cursor, delimiter byte) (*Token, Cursor, bool) {
	newCursor := cursor
	if source[newCursor.pointer] != delimiter {
		return nil, cursor, false
	}

	newCursor.pointer++
	newCursor.loc.col++

	var value []byte

	for ; newCursor.pointer < uint(len(source)); newCursor.pointer++ {
		character := source[newCursor.pointer]
		if character == delimiter {
			if newCursor.pointer+1 >= uint(len(source)) || source[newCursor.pointer+1] != delimiter {
				return NewToken(stringKind, cursor.loc, string(value)), newCursor, true
			} else {
				value = append(value, character)
				newCursor.pointer++
				newCursor.loc.col++
			}
		}
		value = append(value, character)
		newCursor.loc.col++
	}

	return nil, cursor, false
}

func lexString(source string, cursor Cursor) (*Token, Cursor, bool) {
	return lexCharacterDelimited(source, cursor, '\'')
}

func longestMatch(source string, cursor Cursor, options []string) string {
	var value []byte
	var skipList []int
	var match string

	var originCurosr = cursor

	for cursor.pointer < uint(len(source)) {
		value = append(value, bytes.ToLower([]byte{source[cursor.pointer]})...)
		cursor.pointer++
		cursor.loc.col++

		for i, option := range options {
			for _, skip := range skipList {
				if skip == i {
					continue
				}
			}

			if string(value) == option {
				skipList = append(skipList, i)
				if len(option) > len(match) {
					match = option
					continue
				}
			}

			sharePrefix := string(value) == source[:cursor.pointer-originCurosr.pointer]
			tooLong := len(value) >= len(options)
			if !sharePrefix || tooLong {
				skipList = append(skipList, i)
			}
		}

		if len(skipList) == len(options) {
			return match
		}
	}
	return ""
}

// ? Here to skip space
func lexSymbol(source string, cursor Cursor) (*Token, Cursor, bool) {
	character := source[cursor.pointer]
	originCursor := cursor
	cursor.pointer++
	cursor.loc.col++

	switch character {
	case '\n':
		cursor.loc.line++
		cursor.loc.col = 0
		fallthrough
	case '\t':
		fallthrough
	case ' ':
		return nil, cursor, false
	}

	symbols := []symbol{
		commaSymbol,
		leftParenSymbol,
		rightParenSymbol,
		semiColonSymbol,
		asteriskSymbol,
	}

	var options []string
	for _, symbol := range symbols {
		options = append(options, string(symbol))
	}

	match := longestMatch(source, originCursor, options)

	if match == "" {
		return nil, originCursor, false
	}

	cursor.pointer = originCursor.pointer + uint(len(match))
	cursor.loc.col = originCursor.loc.col + len(match)

	return NewToken(symbolKind, originCursor.loc, match), cursor, true

}

func lexKeyword(source string, cursor Cursor) (*Token, Cursor, bool) {
	originCurosr := cursor

	keywords := []keyword{
		selectKeyword,
		fromKeyword,
		insertKeyword,
		intoKeyword,
		intKeyword,
		valuesKeyword,
		createKeyword,
		textKeyword,
		whereKeyword,
	}

	var options []string
	for _, keyword := range keywords {
		options = append(options, string(keyword))
	}

	match := longestMatch(source, originCurosr, options)
	if match == "" {
		return nil, originCurosr, false
	}

	return NewToken(keywordKind, originCurosr.loc, match), cursor, true
}

func lexIdentifier(source string, cursor Cursor) (*Token, Cursor, bool) {
	if token, newCursor, ok := lexCharacterDelimited(source, cursor, '"'); ok {
		return token, newCursor, true
	}

	newCursor := cursor
	character := source[newCursor.pointer]

	isAlphabetical := (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z')

	if !isAlphabetical {
		return nil, cursor, false
	}

	newCursor.pointer++
	newCursor.loc.col++

	var value []byte = []byte{character}
	for newCursor.pointer < uint(len(source)) {
		character := source[newCursor.pointer]
		isAlphabetical := (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z')
		isNumeric := character >= '1' && character <= '9'
		if isAlphabetical || isNumeric || character == '_' || character == '$' {
			value = append(value, character)
			newCursor.pointer++
			newCursor.loc.col++
			continue
		} else {
			break
		}
	}

	return NewToken(identifierKind, cursor.loc, string(value)), newCursor, true
}
