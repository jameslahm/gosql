package lex

type Cursor struct {
	pointer uint
	loc     Location
}

func NewCursor(pointer uint, loc Location) Cursor {
	return Cursor{
		pointer: pointer,
		loc:     loc,
	}
}
