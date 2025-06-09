package utils

import (
	"fmt"
	"reflect"
	"sync"

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

var (
	cache = make(map[string][]reflect.Value)
	mu    sync.Mutex
)

func CallWithCache(fn any, params ...any) []any {
	fnVal := reflect.ValueOf(fn)
	fnPtr := fmt.Sprintf("%#x", fnVal.Pointer())

	key := fnPtr + ":"
	for _, p := range params {
		key += fmt.Sprintf("%#v,", p)
	}

	mu.Lock()
	cachedResult, found := cache[key]
	mu.Unlock()
	if found {
		res := make([]any, len(cachedResult))
		for i, v := range cachedResult {
			res[i] = v.Interface()
		}
		return res
	}

	in := make([]reflect.Value, len(params))
	for i, p := range params {
		in[i] = reflect.ValueOf(p)
	}

	out := fnVal.Call(in)

	mu.Lock()
	cache[key] = out
	mu.Unlock()

	res := make([]any, len(out))
	for i, v := range out {
		res[i] = v.Interface()
	}
	return res
}
