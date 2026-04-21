package store

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	ErrNotFound     = errors.New("container: value not found")
	ErrDuplicate    = errors.New("container: duplicate value")
	ErrInvalidValue = errors.New("container: invalid value")
)

type NotFoundError struct {
	Name string
	Type reflect.Type
}

func (e *NotFoundError) Error() string {
	if e.Name == "" {
		return fmt.Sprintf("container: no value found for type %v", e.Type)
	}
	return fmt.Sprintf("container: no value found for name %q and type %v", e.Name, e.Type)
}

func (e *NotFoundError) Unwrap() error { return ErrNotFound }

type DuplicateError struct {
	Name string
	Type reflect.Type
}

func (e *DuplicateError) Error() string {
	if e.Name == "" {
		return fmt.Sprintf("container: duplicate value for type %v", e.Type)
	}
	return fmt.Sprintf("container: duplicate value for name %q and type %v", e.Name, e.Type)
}

func (e *DuplicateError) Unwrap() error { return ErrDuplicate }
