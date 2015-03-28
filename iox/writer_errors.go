package iox

import (
	. "fmt"
)

/*
	Error raised by WriterFromInterface() when is called with an argument of an unexpected type.
*/
type WriterUnrefinableFromInterface struct {
	wat interface{}
}

func (err WriterUnrefinableFromInterface) Error() string {
	return Sprintf("WriterFromInterface cannot refine type \"%T\" to a Reader", err.wat)
}
