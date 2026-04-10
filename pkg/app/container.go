package app

// container implements the Container interface with support for unnamed bindings,
// named bindings, multi-binding, and reflective invocation.

import (
	"fmt"
	"reflect"
	"sync"
)

// container stores provided values indexed by type and optional name.
type container struct {
	mu     sync.RWMutex
	byType map[reflect.Type][]reflect.Value
	byName map[string]reflect.Value
}

// newContainer creates an empty dependency container.
func newContainer() *container {
	return &container{
		byType: make(map[reflect.Type][]reflect.Value),
		byName: make(map[string]reflect.Value),
	}
}

// Provide stores an unnamed value in the container.
func (c *container) Provide(value any) error {
	if value == nil {
		return fmt.Errorf("app: cannot provide nil value")
	}

	v := reflect.ValueOf(value)
	if isNilValue(v) {
		return fmt.Errorf("app: cannot provide nil value of type %T", value)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.byType[v.Type()] = append(c.byType[v.Type()], v)
	return nil
}

// ProvideNamed stores a named value in the container.
func (c *container) ProvideNamed(name string, value any) error {
	if name == "" {
		return fmt.Errorf("app: named provide requires a non-empty name")
	}
	if value == nil {
		return fmt.Errorf("app: cannot provide nil value")
	}

	v := reflect.ValueOf(value)
	if isNilValue(v) {
		return fmt.Errorf("app: cannot provide nil value of type %T", value)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.byName[name]; exists {
		return fmt.Errorf("app: duplicate named provide for %q", name)
	}
	c.byName[name] = v
	c.byType[v.Type()] = append(c.byType[v.Type()], v)
	return nil
}

// Resolve assigns a single matching value into target.
func (c *container) Resolve(target any) error {
	resolved, err := c.resolveValue(target)
	if err != nil {
		return err
	}
	reflect.ValueOf(target).Elem().Set(resolved)
	return nil
}

// ResolveNamed assigns the named value into target.
func (c *container) ResolveNamed(name string, target any) error {
	if target == nil {
		return fmt.Errorf("app: resolve target must be a non-nil pointer")
	}
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr || targetValue.IsNil() {
		return fmt.Errorf("app: resolve target must be a non-nil pointer")
	}
	targetType := targetValue.Elem().Type()

	c.mu.RLock()
	value, ok := c.byName[name]
	c.mu.RUnlock()
	if !ok {
		return fmt.Errorf("app: no named value found for %q", name)
	}
	if !value.Type().AssignableTo(targetType) {
		return fmt.Errorf("app: named value %q of type %s is not assignable to %s", name, value.Type(), targetType)
	}
	targetValue.Elem().Set(value)
	return nil
}

// MustResolve assigns a single matching value into target and panics on error.
func (c *container) MustResolve(target any) {
	if err := c.Resolve(target); err != nil {
		panic(err)
	}
}

// Invoke resolves function arguments from the container and calls fn.
func (c *container) Invoke(fn any) error {
	if fn == nil {
		return fmt.Errorf("app: invoke requires a non-nil function")
	}
	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		return fmt.Errorf("app: invoke target must be a function")
	}
	t := v.Type()

	args := make([]reflect.Value, 0, t.NumIn())
	for i := 0; i < t.NumIn(); i++ {
		arg := reflect.New(t.In(i))
		resolved, err := c.resolveValue(arg.Interface())
		if err != nil {
			return fmt.Errorf("app: invoke resolve arg %d (%s): %w", i, t.In(i), err)
		}
		args = append(args, resolved)
	}

	results := v.Call(args)
	if len(results) == 0 {
		return nil
	}
	last := results[len(results)-1]
	errType := reflect.TypeOf((*error)(nil)).Elem()
	if last.Type().Implements(errType) && !last.IsNil() {
		return last.Interface().(error)
	}
	return nil
}

// resolveValue finds a single value assignable to target.
func (c *container) resolveValue(target any) (reflect.Value, error) {
	if target == nil {
		return reflect.Value{}, fmt.Errorf("app: resolve target must be a non-nil pointer")
	}
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr || targetValue.IsNil() {
		return reflect.Value{}, fmt.Errorf("app: resolve target must be a non-nil pointer")
	}
	targetType := targetValue.Elem().Type()

	c.mu.RLock()
	defer c.mu.RUnlock()

	if values, ok := c.byType[targetType]; ok {
		if len(values) == 1 {
			return values[0], nil
		}
		if len(values) > 1 {
			return reflect.Value{}, fmt.Errorf("app: multiple values satisfy type %s", targetType)
		}
	}

	var matches []reflect.Value
	for typ, values := range c.byType {
		if !typ.AssignableTo(targetType) {
			continue
		}
		matches = append(matches, values...)
	}
	if len(matches) == 0 {
		return reflect.Value{}, fmt.Errorf("app: no value found for type %s", targetType)
	}
	if len(matches) > 1 {
		return reflect.Value{}, fmt.Errorf("app: multiple values satisfy type %s", targetType)
	}
	return matches[0], nil
}

// isNilValue reports whether v is a nil-able kind with a nil value.
func isNilValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}
