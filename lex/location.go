package lex

type Location struct {
	Line int
	Col  int
}

func NewLocation() Location {
	return Location{
		Line: 0,
		Col:  0,
	}
}
