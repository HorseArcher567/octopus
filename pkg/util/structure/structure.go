package structure

import (
	"github.com/k8s-practice/octopus/pkg/log"
	"reflect"
	"strconv"
	"strings"
)

func Unmarshal(in, out interface{}) error {
	return UnmarshalWithTag(in, out, "")
}

func UnmarshalWithTag(in, out interface{}, tag string) error {
	outValue := reflect.ValueOf(out)
	log.Debugln("outValue:", outValue, outValue.Kind(), outValue.Type().Kind())
	if outValue.Kind() != reflect.Ptr || !outValue.IsValid() || outValue.IsNil() {
		return &InvalidTypeError{Type: reflect.TypeOf(out)}
	}
	for ; outValue.Kind() == reflect.Ptr; outValue = reflect.Indirect(outValue) {
	}
	if outValue.Kind() == reflect.Invalid {
		return &InvalidTypeError{Type: reflect.TypeOf(out)}
	}

	return unmarshal(in, outValue, false, tag)
}

func unmarshal(in interface{}, outValue reflect.Value, strict bool, tag string) error {
	inValue := reflect.ValueOf(in)
	for ; inValue.Kind() == reflect.Ptr; inValue = reflect.Indirect(inValue) {
	}
	if in == nil || !inValue.IsValid() {
		return nil
	}

	switch classifyKind(outValue.Kind()) {
	case reflect.Bool:
		return toBool(inValue, outValue, strict)
	case reflect.String:
		return toString(inValue, outValue, strict)
	case reflect.Int:
		return toInt(inValue, outValue, strict)
	case reflect.Uint:
		return toUint(inValue, outValue, strict)
	case reflect.Float32:
		return toFloat(inValue, outValue, strict)
	case reflect.Slice:
		return toSlice(inValue, outValue, strict, tag)
	case reflect.Array:
		return toArray(inValue, outValue, strict, tag)
	case reflect.Struct:
		return toStruct(inValue, outValue, strict, tag)
	case reflect.Map:
		return toMap(inValue, outValue, strict, tag)
	default:
		return &UnsupportedTypeError{inValue, outValue.Type()}
	}
}

func classifyKind(kind reflect.Kind) reflect.Kind {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return reflect.Int
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return reflect.Uint
	case reflect.Float32, reflect.Float64:
		return reflect.Float32
	default:
		return kind
	}
}

func empty(v reflect.Value) bool {
	switch classifyKind(v.Kind()) {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int:
		return v.Int() == 0
	case reflect.Uint, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

func toBool(inValue, outValue reflect.Value, strict bool) error {
	kind := classifyKind(inValue.Kind())
	switch {
	case kind == reflect.Bool:
		outValue.SetBool(inValue.Bool())
	case kind == reflect.Int && !strict:
		outValue.SetBool(inValue.Int() != 0)
	case kind == reflect.Uint && !strict:
		outValue.SetBool(inValue.Uint() != 0)
	case kind == reflect.Float32 && !strict:
		outValue.SetBool(inValue.Float() != 0)
	case kind == reflect.String && !strict:
		if val, err := strconv.ParseBool(inValue.String()); err != nil {
			return err
		} else {
			outValue.SetBool(val)
		}
	default:
		return &UnsupportedTypeError{inValue, outValue.Type()}
	}
	return nil
}

func toInt(inValue, outValue reflect.Value, strict bool) error {
	kind := classifyKind(inValue.Kind())
	switch {
	case kind == reflect.Bool && !strict:
		if inValue.Bool() {
			outValue.SetInt(1)
		} else {
			outValue.SetInt(0)
		}
	case kind == reflect.Int:
		outValue.SetInt(inValue.Int())
	case kind == reflect.Uint && !strict:
		outValue.SetInt(int64(inValue.Uint()))
	case kind == reflect.Float32 && !strict:
		outValue.SetInt(int64(inValue.Float()))
	case kind == reflect.String && !strict:
		val, err := strconv.ParseInt(inValue.String(), 0, outValue.Type().Bits())
		if err == nil {
			outValue.SetInt(val)
		} else {
			return &UnsupportedTypeError{inValue, outValue.Type()}
		}
	default:
		return &UnsupportedTypeError{inValue, outValue.Type()}
	}
	return nil
}

func toUint(inValue, outValue reflect.Value, strict bool) error {
	kind := classifyKind(inValue.Kind())
	switch {
	case kind == reflect.Bool && !strict:
		if inValue.Bool() {
			outValue.SetUint(1)
		} else {
			outValue.SetUint(0)
		}
	case kind == reflect.Int && !strict:
		outValue.SetUint(uint64(inValue.Int()))
	case kind == reflect.Uint:
		outValue.SetUint(inValue.Uint())
	case kind == reflect.Float32 && !strict:
		outValue.SetUint(uint64(inValue.Float()))
	case kind == reflect.String && !strict:
		val, err := strconv.ParseUint(inValue.String(), 0, outValue.Type().Bits())
		if err == nil {
			outValue.SetUint(val)
		} else {
			return &UnsupportedTypeError{inValue, outValue.Type()}
		}
	default:
		return &UnsupportedTypeError{inValue, outValue.Type()}
	}
	return nil
}

func toFloat(inValue, outValue reflect.Value, strict bool) error {
	kind := classifyKind(inValue.Kind())
	switch {
	case kind == reflect.Bool && !strict:
		if inValue.Bool() {
			outValue.SetFloat(1)
		} else {
			outValue.SetFloat(0)
		}
	case kind == reflect.Int && !strict:
		outValue.SetFloat(float64(inValue.Int()))
	case kind == reflect.Uint && !strict:
		outValue.SetFloat(float64(inValue.Uint()))
	case kind == reflect.Float32:
		outValue.SetFloat(inValue.Float())
	case kind == reflect.String && !strict:
		val, err := strconv.ParseFloat(inValue.String(), outValue.Type().Bits())
		if err == nil {
			outValue.SetFloat(val)
		} else {
			return &UnsupportedTypeError{inValue, outValue.Type()}
		}
	default:
		return &UnsupportedTypeError{inValue, outValue.Type()}
	}
	return nil
}

func toString(inValue, outValue reflect.Value, strict bool) error {
	kind := classifyKind(inValue.Kind())
	switch {
	case kind == reflect.Bool && !strict:
		if inValue.Bool() {
			outValue.SetString("1")
		} else {
			outValue.SetString("0")
		}
	case kind == reflect.Int && !strict:
		outValue.SetString(strconv.FormatInt(inValue.Int(), 10))
	case kind == reflect.Uint && !strict:
		outValue.SetString(strconv.FormatUint(inValue.Uint(), 10))
	case kind == reflect.Float32 && !strict:
		outValue.SetString(strconv.FormatFloat(inValue.Float(), 'f', -1, 64))
	case kind == reflect.String:
		outValue.SetString(inValue.String())
	default:
		return &UnsupportedTypeError{inValue, outValue.Type()}
	}
	return nil
}

func toSlice(inValue, outValue reflect.Value, strict bool, tag string) error {
	if inValue.Kind() != reflect.Slice && inValue.Kind() != reflect.Array {
		return &UnsupportedTypeError{inValue, outValue.Type()}
	}

	sliceType := reflect.SliceOf(outValue.Type().Elem())
	sliceValue := outValue
	length := inValue.Len()
	if sliceValue.IsNil() {
		sliceValue = reflect.MakeSlice(sliceType, length, length)
	}
	for sliceValue.Len() < inValue.Len() {
		sliceValue = reflect.Append(sliceValue, reflect.Zero(outValue.Type().Elem()))
	}
	for i := 0; i < length; i++ {
		err := unmarshal(inValue.Index(i).Interface(), sliceValue.Index(i), strict, tag)
		if err != nil {
			return err
		}
	}

	outValue.Set(sliceValue)
	return nil
}

func toArray(inValue, outValue reflect.Value, strict bool, tag string) error {
	if inValue.Kind() != reflect.Slice && inValue.Kind() != reflect.Array {
		return &UnsupportedTypeError{inValue, outValue.Type()}
	}
	if inValue.Kind() == reflect.Slice && inValue.Len() != outValue.Len() {
		return &UnsupportedTypeError{inValue, outValue.Type()}
	}

	for i := 0; i < inValue.Len(); i++ {
		err := unmarshal(inValue.Index(i).Interface(), outValue.Index(i), strict, tag)
		if err != nil {
			return err
		}
	}

	return nil
}

func toStruct(inValue, outValue reflect.Value, strict bool, tag string) error {
	switch inValue.Kind() {
	case reflect.Map:
		return mapToStruct(inValue, outValue, strict, tag)
	case reflect.Struct:
		mapType := reflect.TypeOf((map[string]interface{})(nil))
		mapValue := reflect.MakeMap(mapType)
		if err := structToMap(inValue, mapValue, strict, tag); err != nil {
			return mapToStruct(mapValue, outValue, strict, tag)
		} else {
			return err
		}
	default:
		return &UnsupportedTypeError{inValue, outValue.Type()}
	}
}

func toMap(inValue, outValue reflect.Value, strict bool, tag string) error {
	switch inValue.Kind() {
	case reflect.Struct:
		return structToMap(inValue, outValue, strict, tag)
	default:
		return &UnsupportedTypeError{inValue, outValue.Type()}
	}
}

func structToMap(inValue, outValue reflect.Value, strict bool, tag string) error {
	mapValue := reflect.MakeMap(outValue.Type())

	inType := inValue.Type()
	for i := 0; i < inType.NumField(); i++ {
		structField := inType.Field(i)
		if structField.PkgPath != "" {
			continue
		}

		filedValue := inValue.Field(i)
		if !filedValue.Type().AssignableTo(outValue.Type().Elem()) {
			return &UnsupportedTypeError{filedValue, outValue.Type().Elem()}
		}

		tagValue := structField.Tag.Get(tag)
		name, omitempty := parseTag(tagValue)
		if omitempty && empty(filedValue) {
			continue
		}
		if name == "" {
			name = structField.Name
		}

		switch filedValue.Kind() {
		case reflect.Struct:
			outType := outValue.Type()
			mapType := reflect.MapOf(outType.Key(), outType.Elem())
			elemMap := reflect.MakeMap(mapType)
			mapPtr := reflect.New(elemMap.Type())
			reflect.Indirect(mapPtr).Set(elemMap)

			if err := unmarshal(filedValue.Interface(), reflect.Indirect(mapPtr), strict, tag); err != nil {
				return err
			}
			mapValue.SetMapIndex(reflect.ValueOf(name), reflect.Indirect(mapPtr))
		default:
			mapValue.SetMapIndex(reflect.ValueOf(name), filedValue)
		}
	}

	outValue.Set(mapValue)
	return nil
}

func parseTag(tag string) (name string, omitempty bool) {
	fields := strings.Split(tag, ",")
	if len(fields) == 0 {
		return
	}

	name = fields[0]
	for _, flag := range fields[1:] {
		switch flag {
		case "omitempty":
			omitempty = true
		}
	}

	return
}

func mapToStruct(inValue, outValue reflect.Value, strict bool, tag string) error {
	keyType := inValue.Type().Key()
	if keyType.Kind() != reflect.String && keyType.Kind() != reflect.Interface {
		return &UnsupportedTypeError{inValue, outValue.Type()}
	}

	outType := outValue.Type()
	for i := 0; i < outType.NumField(); i++ {
		structField := outType.Field(i)
		fieldValue := outValue.Field(i)
		if fieldValue.Kind() == reflect.Ptr && fieldValue.Elem().Kind() == reflect.Struct {
			fieldValue = fieldValue.Elem()
		}

		tagValue := structField.Tag.Get(tag)
		alias, _ := parseTag(tagValue)
		if alias == "-" {
			continue
		}

		elemValue := inValue.MapIndex(reflect.ValueOf(alias))
		if alias != "" && elemValue.IsValid() {
			if err := unmarshal(elemValue.Interface(), fieldValue, strict, tag); err != nil {
				return err
			}
			continue
		}

		elemValue = inValue.MapIndex(reflect.ValueOf(structField.Name))
		if elemValue.IsValid() {
			if err := unmarshal(elemValue.Interface(), fieldValue, strict, tag); err != nil {
				return err
			}
			continue
		}

		elemValue = inValue.MapIndex(reflect.ValueOf(strings.ToLower(structField.Name)))
		if elemValue.IsValid() {
			if err := unmarshal(elemValue.Interface(), fieldValue, strict, tag); err != nil {
				return err
			}
			continue
		}
	}

	return nil
}
