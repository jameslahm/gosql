package lex

type Location struct {
	line int
	col  int
}

func NewLocation() Location {
	return Location{
		line: 0,
		col:  0,
	}
}
