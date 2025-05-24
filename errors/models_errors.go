package errors

import "fmt"

type ErrValidation struct {
	StructName     string
	DerivedFromErr error
}

func (e ErrValidation) Error() string {
	return fmt.Sprintf("error, validation error for struct: %s derived from: %s",
		e.StructName, e.DerivedFromErr.Error())
}
