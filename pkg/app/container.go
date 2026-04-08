package app

import (
	"fmt"
	"reflect"
)

type container struct {
	values map[reflect.Type]reflect.Value
}

func newContainer() *container {
	return &container{values: make(map[reflect.Type]reflect.Value)}
}

func (c *container) Provide(value any) error {
	if value == nil {
		return fmt.Errorf("app: cannot provide nil value")
	}

	v := reflect.ValueOf(value)
	if isNilValue(v) {
		return fmt.Errorf("app: cannot provide nil value of type %T", value)
	}

	t := v.Type()
	if _, exists := c.values[t]; exists {
		return fmt.Errorf("app: duplicate provide for type %s", t)
	}

	c.values[t] = v
	return nil
}

func (c *container) Resolve(target any) error {
	if target == nil {
		return fmt.Errorf("app: resolve target must be a non-nil pointer")
	}

	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr || targetValue.IsNil() {
		return fmt.Errorf("app: resolve target must be a non-nil pointer")
	}

	targetType := targetValue.Elem().Type()

	if value, ok := c.values[targetType]; ok {
		targetValue.Elem().Set(value)
		return nil
	}

	var (
		match     reflect.Value
		matchType reflect.Type
		found     bool
	)
	for typ, value := range c.values {
		if !typ.AssignableTo(targetType) {
			continue
		}
		if found {
			return fmt.Errorf("app: multiple values satisfy type %s: %s and %s", targetType, matchType, typ)
		}
		match = value
		matchType = typ
		found = true
	}
	if !found {
		return fmt.Errorf("app: no value found for type %s", targetType)
	}

	targetValue.Elem().Set(match)
	return nil
}

func (c *container) MustResolve(target any) {
	if err := c.Resolve(target); err != nil {
		panic(err)
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
