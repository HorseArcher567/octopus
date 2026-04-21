package store

import (
	"fmt"
	"reflect"
)

func Get[T any](s Store) (T, error) {
	return GetNamed[T](s, "")
}

func GetNamed[T any](s Store, name string) (T, error) {
	var zero T
	typ := reflect.TypeFor[T]()
	v, err := s.GetNamed(name, typ)
	if err != nil {
		return zero, err
	}
	typed, ok := v.(T)
	if !ok {
		return zero, fmt.Errorf("store: internal type mismatch for name %q and type %v", name, typ)
	}
	return typed, nil
}

func MustGet[T any](s Store) T {
	v, err := Get[T](s)
	if err != nil {
		panic(err)
	}
	return v
}

func MustGetNamed[T any](s Store, name string) T {
	v, err := GetNamed[T](s, name)
	if err != nil {
		panic(err)
	}
	return v
}
