package store

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

type key struct {
	name string
	typ  reflect.Type
}

type entry struct {
	value   any
	closeFn func() error
}

type Store interface {
	Set(value any, opts ...SetOption) error
	SetNamed(name string, value any, opts ...SetOption) error
	Get(typ reflect.Type) (any, error)
	GetNamed(name string, typ reflect.Type) (any, error)
	Close() error
}

type store struct {
	mu      sync.RWMutex
	entries map[key]entry
}

func New() Store {
	return &store{entries: make(map[key]entry)}
}

func (s *store) Set(value any, opts ...SetOption) error {
	return s.SetNamed("", value, opts...)
}

func (s *store) SetNamed(name string, value any, opts ...SetOption) error {
	if value == nil {
		return fmt.Errorf("%w: value cannot be nil", ErrInvalidValue)
	}
	v := reflect.ValueOf(value)
	if isNilValue(v) {
		return fmt.Errorf("%w: value of type %T cannot be nil", ErrInvalidValue, value)
	}

	e := entry{value: value}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(&e); err != nil {
			return err
		}
	}

	k := key{name: name, typ: v.Type()}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.entries[k]; exists {
		return &DuplicateError{Name: name, Type: v.Type()}
	}
	s.entries[k] = e
	return nil
}

func (s *store) Get(typ reflect.Type) (any, error) {
	return s.GetNamed("", typ)
}

func (s *store) GetNamed(name string, typ reflect.Type) (any, error) {
	if typ == nil {
		return nil, fmt.Errorf("%w: type cannot be nil", ErrInvalidValue)
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[key{name: name, typ: typ}]
	if !ok {
		return nil, &NotFoundError{Name: name, Type: typ}
	}
	return e.value, nil
}

func (s *store) Close() error {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	entries := make([]entry, 0, len(s.entries))
	for _, e := range s.entries {
		entries = append(entries, e)
	}
	s.mu.RUnlock()

	var errs []error
	for _, e := range entries {
		if e.closeFn == nil {
			continue
		}
		if err := e.closeFn(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func isNilValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}
