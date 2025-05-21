package errors

import (
	"fmt"
)

type ErrConfigLoadingError struct {
	DerivedFromErr error
}

type ErrConfigNotValid struct {
	DerivedFromErr error
}

func (e *ErrConfigLoadingError) Error() string {
	return fmt.Sprintf("error loading configuration for helga please check the configuration again derived error: %s", e.DerivedFromErr.Error())
}

func (e *ErrConfigNotValid) Error() string {
	return fmt.Sprintf("error validating configuration for helga please check the configuration again derived error: %s", e.DerivedFromErr.Error())
}
