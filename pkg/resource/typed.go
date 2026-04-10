package resource

import "fmt"

type getter interface {
	Get(kind, name string) (any, error)
}

type mustGetter interface {
	MustGet(kind, name string) any
}

// As resolves a typed resource from a generic runtime.
func As[T any](rt getter, kind, name string) (T, error) {
	var zero T
	v, err := rt.Get(kind, name)
	if err != nil {
		return zero, err
	}
	typed, ok := v.(T)
	if !ok {
		return zero, fmt.Errorf("resource: %s[%s] has type %T", kind, name, v)
	}
	return typed, nil
}

// MustAs resolves a typed resource or panics.
func MustAs[T any](rt mustGetter, kind, name string) T {
	v := rt.MustGet(kind, name)
	typed, ok := v.(T)
	if !ok {
		panic(fmt.Sprintf("resource: %s[%s] has type %T", kind, name, v))
	}
	return typed
}
