package iox

import (
	. "fmt"
)

/*
	Error raised by ReaderFromInterface() when is called with an argument of an unexpected type.
*/
type ReaderUnrefinableFromInterface struct {
	wat interface{}
}

func (err ReaderUnrefinableFromInterface) Error() string {
	return Sprintf("ReaderFromInterface cannot refine type \"%T\" to a Reader", err.wat)
}
