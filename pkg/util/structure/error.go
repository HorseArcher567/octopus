package structure

import (
	"fmt"
	"reflect"
)

// An InvalidTypeError describes an invalid output argument.
// (The output argument to structure must be a non-nil pointer.)
type InvalidTypeError struct {
	Type reflect.Type
}

func (e *InvalidTypeError) Error() string {
	if e.Type == nil {
		return "structure: To(nil)"
	}

	if e.Type.Kind() != reflect.Ptr {
		return "structure: To(non-pointer " + e.Type.String() + ")"
	}
	return "structure: To(nil " + e.Type.String() + ")"
}

type UnsupportedTypeError struct {
	InValue reflect.Value
	OutType reflect.Type
}

func (e *UnsupportedTypeError) Error() string {
	inType, outType := "nil", "nil"
	if e.InValue.Type() != nil {
		inType = e.InValue.Type().String()
	}
	if e.OutType != nil {
		outType = e.OutType.String()
	}

	return fmt.Sprintf("structure: Unsupported(%s:%v to %s", inType, e.InValue, outType)
}
