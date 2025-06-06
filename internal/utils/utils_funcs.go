package utils

import (
	"fmt"

	helga_errors "github.com/fennet82/helga/pkg/errors"
)

type Validatable interface {
	Validate() []error
	String() string
}

func ToValidatableSlice[T Validatable](ss []T) []Validatable {
	result := make([]Validatable, len(ss))

	for i, s := range ss {
		result[i] = s
	}

	return result
}

func FromValidatableSlice[T any](vals []Validatable) []T {
	result := make([]T, 0, len(vals))
	for _, v := range vals {
		typed, ok := v.(T)
		if !ok {
			helga_errors.HandleError(fmt.Errorf("type assertion failed for value: %+v", v))
		}
		result = append(result, typed)
	}
	return result
}

func FilterByErrFunc[T any](ss []T, f func(T) error) (ret []T) {
	for _, s := range ss {
		if f(s) == nil {
			ret = append(ret, s)
		}
	}

	return
}

func FilterByValidation(ss []Validatable, vErrMsg string) (errs []error, ret []Validatable) {
	for _, s := range ss {
		if s.Validate() == nil {
			ret = append(ret, s)
		} else {
			errs = append(errs, fmt.Errorf(vErrMsg, s.String()))
		}
	}

	return
}
