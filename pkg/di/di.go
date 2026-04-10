// Package di provides dependency binding, resolution, and invocation helpers.
package di

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"sync"
)

var (
	// ErrNotFound reports that no binding matched the requested lookup.
	ErrNotFound = errors.New("di: binding not found")
	// ErrAmbiguous reports that multiple bindings matched a single-value lookup.
	ErrAmbiguous = errors.New("di: binding is ambiguous")
	// ErrInvalidTarget reports that a resolve target is invalid.
	ErrInvalidTarget = errors.New("di: invalid target")
	// ErrInvalidValue reports that a provided value is invalid.
	ErrInvalidValue = errors.New("di: invalid value")
	// ErrInvalidFunc reports that an invoke target is invalid.
	ErrInvalidFunc = errors.New("di: invalid function")
)

// Binder exposes dependency publication.
type Binder interface {
	Provide(value any) error
	ProvideNamed(name string, value any) error
}

// Resolver exposes dependency lookup.
type Resolver interface {
	Resolve(target any) error
	ResolveNamed(name string, target any) error
	ResolveAll(target any) error
	ResolveAllNamed(name string, target any) error
}

// Invoker exposes reflective function invocation with auto-resolved arguments.
type Invoker interface {
	Invoke(fn any) error
}

// Container exposes dependency publication, resolution, and invocation.
type Container interface {
	Binder
	Resolver
	Invoker
}

// LookupError describes a failed dependency lookup.
type LookupError struct {
	Kind  error
	Name  string
	Type  reflect.Type
	Count int
}

func (e *LookupError) Error() string {
	if e.Name == "" {
		switch e.Kind {
		case ErrNotFound:
			return fmt.Sprintf("di: no binding found for type %v", e.Type)
		case ErrAmbiguous:
			return fmt.Sprintf("di: %d bindings found for type %v", e.Count, e.Type)
		}
	}
	switch e.Kind {
	case ErrNotFound:
		return fmt.Sprintf("di: no binding found for name %q and type %v", e.Name, e.Type)
	case ErrAmbiguous:
		return fmt.Sprintf("di: %d bindings found for name %q and type %v", e.Count, e.Name, e.Type)
	default:
		return "di: lookup error"
	}
}

func (e *LookupError) Unwrap() error { return e.Kind }

type binding struct {
	value reflect.Value
	seq   uint64
}

type namespace map[reflect.Type][]binding

type container struct {
	mu      sync.RWMutex
	byName  map[string]namespace
	nextSeq uint64
}

// New creates an empty dependency container.
func New() Container {
	return &container{byName: make(map[string]namespace)}
}

func (c *container) Provide(value any) error {
	return c.provide("", value)
}

func (c *container) ProvideNamed(name string, value any) error {
	if name == "" {
		return fmt.Errorf("%w: named provide requires a non-empty name", ErrInvalidValue)
	}
	return c.provide(name, value)
}

func (c *container) provide(name string, value any) error {
	if value == nil {
		return fmt.Errorf("%w: value cannot be nil", ErrInvalidValue)
	}
	v := reflect.ValueOf(value)
	if isNilValue(v) {
		return fmt.Errorf("%w: value of type %T cannot be nil", ErrInvalidValue, value)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	ns := c.byName[name]
	if ns == nil {
		ns = make(namespace)
		c.byName[name] = ns
	}
	seq := c.nextSeq
	c.nextSeq++
	ns[v.Type()] = append(ns[v.Type()], binding{value: v, seq: seq})
	return nil
}

func (c *container) Resolve(target any) error {
	return c.resolve("", target)
}

func (c *container) ResolveNamed(name string, target any) error {
	return c.resolve(name, target)
}

func (c *container) resolve(name string, target any) error {
	elem, targetType, err := derefResolveTarget(target)
	if err != nil {
		return err
	}
	resolved, err := c.resolveValue(name, targetType)
	if err != nil {
		return err
	}
	elem.Set(resolved)
	return nil
}

func (c *container) ResolveAll(target any) error {
	return c.resolveAll("", target)
}

func (c *container) ResolveAllNamed(name string, target any) error {
	return c.resolveAll(name, target)
}

func (c *container) resolveAll(name string, target any) error {
	sliceValue, elemType, err := derefResolveAllTarget(target)
	if err != nil {
		return err
	}
	matches := c.findMatches(name, elemType)
	result := reflect.MakeSlice(sliceValue.Type(), 0, len(matches))
	for _, m := range matches {
		result = reflect.Append(result, m.value)
	}
	sliceValue.Set(result)
	return nil
}

func (c *container) Invoke(fn any) error {
	fnValue, fnType, err := validateInvokeFunc(fn)
	if err != nil {
		return err
	}

	args := make([]reflect.Value, 0, fnType.NumIn())
	for i := 0; i < fnType.NumIn(); i++ {
		argType := fnType.In(i)
		argValue, err := c.resolveValue("", argType)
		if err != nil {
			return fmt.Errorf("di: invoke resolve arg %d (%s): %w", i, argType, err)
		}
		args = append(args, argValue)
	}

	results := fnValue.Call(args)
	if fnType.NumOut() == 1 {
		if errVal := results[0]; !errVal.IsNil() {
			return errVal.Interface().(error)
		}
	}
	return nil
}

func (c *container) resolveValue(name string, targetType reflect.Type) (reflect.Value, error) {
	matches := c.findMatches(name, targetType)
	switch len(matches) {
	case 0:
		return reflect.Value{}, &LookupError{Kind: ErrNotFound, Name: name, Type: targetType, Count: 0}
	case 1:
		return matches[0].value, nil
	default:
		return reflect.Value{}, &LookupError{Kind: ErrAmbiguous, Name: name, Type: targetType, Count: len(matches)}
	}
}

func (c *container) findMatches(name string, targetType reflect.Type) []binding {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ns := c.byName[name]
	if ns == nil {
		return nil
	}

	var exact []binding
	var assignable []binding
	for providedType, bindings := range ns {
		switch {
		case providedType == targetType:
			exact = append(exact, bindings...)
		case providedType.AssignableTo(targetType):
			assignable = append(assignable, bindings...)
		case targetType.Kind() == reflect.Interface && providedType.Implements(targetType):
			assignable = append(assignable, bindings...)
		}
	}
	matches := exact
	if len(matches) == 0 {
		matches = assignable
	}
	sort.Slice(matches, func(i, j int) bool { return matches[i].seq < matches[j].seq })
	return matches
}

func derefResolveTarget(target any) (reflect.Value, reflect.Type, error) {
	if target == nil {
		return reflect.Value{}, nil, fmt.Errorf("%w: target cannot be nil", ErrInvalidTarget)
	}
	rv := reflect.ValueOf(target)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return reflect.Value{}, nil, fmt.Errorf("%w: target must be a non-nil pointer", ErrInvalidTarget)
	}
	elem := rv.Elem()
	if !elem.CanSet() {
		return reflect.Value{}, nil, fmt.Errorf("%w: target element cannot be set", ErrInvalidTarget)
	}
	return elem, elem.Type(), nil
}

func derefResolveAllTarget(target any) (reflect.Value, reflect.Type, error) {
	if target == nil {
		return reflect.Value{}, nil, fmt.Errorf("%w: target cannot be nil", ErrInvalidTarget)
	}
	rv := reflect.ValueOf(target)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return reflect.Value{}, nil, fmt.Errorf("%w: target must be a non-nil pointer", ErrInvalidTarget)
	}
	elem := rv.Elem()
	if elem.Kind() != reflect.Slice {
		return reflect.Value{}, nil, fmt.Errorf("%w: target must point to a slice", ErrInvalidTarget)
	}
	if !elem.CanSet() {
		return reflect.Value{}, nil, fmt.Errorf("%w: target slice cannot be set", ErrInvalidTarget)
	}
	return elem, elem.Type().Elem(), nil
}

func validateInvokeFunc(fn any) (reflect.Value, reflect.Type, error) {
	if fn == nil {
		return reflect.Value{}, nil, fmt.Errorf("%w: function cannot be nil", ErrInvalidFunc)
	}
	rv := reflect.ValueOf(fn)
	rt := rv.Type()
	if rt.Kind() != reflect.Func {
		return reflect.Value{}, nil, fmt.Errorf("%w: expected function, got %T", ErrInvalidFunc, fn)
	}
	errType := reflect.TypeOf((*error)(nil)).Elem()
	switch rt.NumOut() {
	case 0:
		return rv, rt, nil
	case 1:
		if !rt.Out(0).Implements(errType) {
			return reflect.Value{}, nil, fmt.Errorf("%w: function return type must be error", ErrInvalidFunc)
		}
		return rv, rt, nil
	default:
		return reflect.Value{}, nil, fmt.Errorf("%w: function may return at most one value", ErrInvalidFunc)
	}
}

func isNilValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}
