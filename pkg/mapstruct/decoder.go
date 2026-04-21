package mapstruct

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"time"
)

// Error values returned by the decoder.
var (
	ErrArrayLengthMismatch = errors.New("array length mismatch")
)

// Decoder maps map[string]any values into Go structs.
type Decoder struct {
	// TagName controls which struct tag is used for field lookup.
	// New() currently defaults to the "yaml" tag.
	TagName string
	// StrictMode returns field conversion errors instead of skipping them.
	StrictMode bool
	// TimeLayout controls string-to-time parsing. The default is RFC3339.
	TimeLayout string
}

// New creates a decoder.
func New() *Decoder {
	return &Decoder{
		TagName:    "yaml",
		StrictMode: false,
		TimeLayout: time.RFC3339,
	}
}

// WithTagName sets the struct tag name used during field lookup.
func (d *Decoder) WithTagName(tagName string) *Decoder {
	d.TagName = tagName
	return d
}

// WithStrictMode enables or disables strict decoding.
func (d *Decoder) WithStrictMode(strict bool) *Decoder {
	d.StrictMode = strict
	return d
}

// WithTimeLayout sets the layout used for parsing time strings.
func (d *Decoder) WithTimeLayout(layout string) *Decoder {
	d.TimeLayout = layout
	return d
}

// Decode decodes input into target.
func (d *Decoder) Decode(input map[string]any, target interface{}) error {
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer")
	}
	if targetValue.IsNil() {
		return fmt.Errorf("target must be a non-nil pointer")
	}

	targetValue = d.resolveToStructValue(targetValue)
	if !targetValue.IsValid() || targetValue.Kind() != reflect.Struct {
		return fmt.Errorf("target must resolve to a struct via pointer(s)")
	}

	targetType := targetValue.Type()
	return d.decodeStruct(input, targetValue, targetType)
}

func (d *Decoder) resolveToStructValue(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			if !v.CanSet() {
				return reflect.Value{}
			}
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	return v
}

// decodeStruct decodes a struct value.
func (d *Decoder) decodeStruct(input map[string]any, targetValue reflect.Value, targetType reflect.Type) error {
	for i := 0; i < targetType.NumField(); i++ {
		field := targetType.Field(i)
		fieldValue := targetValue.Field(i)

		// Skip fields that cannot be set.
		if !fieldValue.CanSet() {
			continue
		}

		// Choose the decoding path based on field shape.
		var err error
		if field.Anonymous {
			err = d.decodeEmbeddedField(input, fieldValue, field)
		} else {
			err = d.decodeNormalField(input, fieldValue, field)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// decodeEmbeddedField decodes an embedded field.
func (d *Decoder) decodeEmbeddedField(input map[string]any, fieldValue reflect.Value, field reflect.StructField) error {
	if err := d.decodeStruct(input, fieldValue, field.Type); err != nil {
		return d.handleDecodeError(err, field.Name)
	}
	return nil
}

// decodeNormalField decodes a non-embedded field.
func (d *Decoder) decodeNormalField(input map[string]any, fieldValue reflect.Value, field reflect.StructField) error {
	// Resolve the source field name.
	fieldName := d.getFieldName(field)
	if fieldName == "" {
		return nil
	}

	// Read the input value.
	inputValue, exists := input[fieldName]
	if !exists {
		return nil
	}

	// Decode the field value.
	if err := d.decodeField(inputValue, fieldValue); err != nil {
		return d.handleDecodeError(err, fieldName)
	}

	return nil
}

// handleDecodeError decides whether a field error should be returned.
func (d *Decoder) handleDecodeError(err error, fieldName string) error {
	// Array length mismatches are always returned.
	if errors.Is(err, ErrArrayLengthMismatch) {
		return fmt.Errorf("failed to decode field %s: %w", fieldName, err)
	}

	// In strict mode, return all field errors.
	if d.StrictMode {
		return fmt.Errorf("failed to decode field %s: %w", fieldName, err)
	}

	// In non-strict mode, ignore field errors.
	return nil
}

// getFieldName resolves the source name for a struct field.
func (d *Decoder) getFieldName(field reflect.StructField) string {
	// Without a tag name, fall back to the struct field name.
	if d.TagName == "" {
		return field.Name
	}

	// Read the configured struct tag.
	tag := field.Tag.Get(d.TagName)
	if tag == "" {
		return field.Name
	}

	// A tag value of "-" means the field is ignored.
	if tag == "-" {
		return ""
	}

	return tag
}

// decodeField decodes a single field value.
func (d *Decoder) decodeField(inputValue any, targetValue reflect.Value) error {
	targetType := targetValue.Type()

	// Ignore nil input values.
	if inputValue == nil {
		return nil
	}

	// Fast path for exact type matches.
	if reflect.TypeOf(inputValue) == targetType {
		targetValue.Set(reflect.ValueOf(inputValue))
		return nil
	}

	// Handle pointer targets.
	if targetType.Kind() == reflect.Ptr {
		return d.decodeToPointer(inputValue, targetValue, targetType)
	}

	// Handle slice targets.
	if targetType.Kind() == reflect.Slice {
		return d.decodeToSlice(inputValue, targetValue, targetType)
	}

	// Handle array targets.
	if targetType.Kind() == reflect.Array {
		return d.decodeToArray(inputValue, targetValue, targetType)
	}

	// Handle map targets.
	if targetType.Kind() == reflect.Map {
		return d.decodeToMap(inputValue, targetValue, targetType)
	}

	// Handle struct targets.
	if targetType.Kind() == reflect.Struct {
		// Special-case time.Time.
		if targetType.String() == "time.Time" {
			return d.decodeToTime(inputValue, targetValue)
		}
		return d.decodeToStruct(inputValue, targetValue, targetType)
	}

	// Special-case time.Duration.
	if targetType.String() == "time.Duration" {
		return d.decodeToDuration(inputValue, targetValue)
	}

	// Handle basic scalar types.
	return d.decodeBasicType(inputValue, targetValue, targetType)
}

// decodeToPointer decodes into a pointer target.
func (d *Decoder) decodeToPointer(inputValue any, targetValue reflect.Value, targetType reflect.Type) error {
	elemType := targetType.Elem()
	elemValue := reflect.New(elemType).Elem()

	if err := d.decodeField(inputValue, elemValue); err != nil {
		return err
	}

	targetValue.Set(elemValue.Addr())
	return nil
}

// decodeToSlice decodes into a slice target.
func (d *Decoder) decodeToSlice(inputValue any, targetValue reflect.Value, targetType reflect.Type) error {
	inputSlice, ok := d.toSlice(inputValue)
	if !ok {
		return fmt.Errorf("cannot decode %T to slice", inputValue)
	}

	elemType := targetType.Elem()
	slice := reflect.MakeSlice(targetType, len(inputSlice), len(inputSlice))

	for i, item := range inputSlice {
		elemValue := reflect.New(elemType).Elem()
		if err := d.decodeField(item, elemValue); err != nil {
			return fmt.Errorf("failed to decode slice element %d: %w", i, err)
		}
		slice.Index(i).Set(elemValue)
	}

	targetValue.Set(slice)
	return nil
}

// decodeToArray decodes into an array target.
func (d *Decoder) decodeToArray(inputValue any, targetValue reflect.Value, targetType reflect.Type) error {
	inputSlice, ok := d.toSlice(inputValue)
	if !ok {
		return fmt.Errorf("cannot decode %T to array", inputValue)
	}

	arrayLen := targetType.Len()
	if len(inputSlice) != arrayLen {
		return fmt.Errorf("%w: expected %d, got %d", ErrArrayLengthMismatch, arrayLen, len(inputSlice))
	}

	elemType := targetType.Elem()
	for i, item := range inputSlice {
		elemValue := reflect.New(elemType).Elem()
		if err := d.decodeField(item, elemValue); err != nil {
			return fmt.Errorf("failed to decode array element %d: %w", i, err)
		}
		targetValue.Index(i).Set(elemValue)
	}

	return nil
}

// decodeToMap decodes into a map target. Only string keys are supported.
func (d *Decoder) decodeToMap(inputValue any, targetValue reflect.Value, targetType reflect.Type) error {
	inputMap, ok := d.toMap(inputValue)
	if !ok {
		return fmt.Errorf("cannot decode %T to map", inputValue)
	}

	if targetType.Key().Kind() != reflect.String {
		return fmt.Errorf("unsupported map key type: %s", targetType.Key().Kind())
	}

	if targetValue.IsNil() {
		targetValue.Set(reflect.MakeMapWithSize(targetType, len(inputMap)))
	}

	elemType := targetType.Elem()
	for key, item := range inputMap {
		elemValue := reflect.New(elemType).Elem()
		if err := d.decodeField(item, elemValue); err != nil {
			return fmt.Errorf("failed to decode map value for key %q: %w", key, err)
		}
		targetValue.SetMapIndex(reflect.ValueOf(key), elemValue)
	}

	return nil
}

// decodeToStruct decodes into a struct target.
func (d *Decoder) decodeToStruct(inputValue any, targetValue reflect.Value, targetType reflect.Type) error {
	inputMap, ok := inputValue.(map[string]any)
	if !ok {
		return fmt.Errorf("cannot decode %T to struct", inputValue)
	}

	return d.decodeStruct(inputMap, targetValue, targetType)
}

// decodeToTime decodes into time.Time.
func (d *Decoder) decodeToTime(inputValue any, targetValue reflect.Value) error {
	var t time.Time
	var err error

	switch v := inputValue.(type) {
	case string:
		// Try a small set of supported time layouts.
		formats := []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02 15:04:05Z07:00",
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05",
			"2006-01-02",
		}

		for _, format := range formats {
			t, err = time.Parse(format, v)
			if err == nil {
				targetValue.Set(reflect.ValueOf(t))
				return nil
			}
		}
		return fmt.Errorf("cannot parse time string: %s", v)

	case time.Time:
		targetValue.Set(reflect.ValueOf(v))
		return nil

	case int64:
		// Unix timestamp (seconds)
		t = time.Unix(v, 0)
		targetValue.Set(reflect.ValueOf(t))
		return nil

	case float64:
		// Unix timestamp, optionally with fractional seconds.
		sec := int64(v)
		nsec := int64((v - float64(sec)) * 1e9)
		t = time.Unix(sec, nsec)
		targetValue.Set(reflect.ValueOf(t))
		return nil

	default:
		return fmt.Errorf("cannot decode %T to time.Time", inputValue)
	}
}

// decodeToDuration decodes into time.Duration.
func (d *Decoder) decodeToDuration(inputValue any, targetValue reflect.Value) error {
	var duration time.Duration
	var err error

	switch v := inputValue.(type) {
	case string:
		// Parse duration strings such as "10s", "5m", or "1h".
		duration, err = time.ParseDuration(v)
		if err != nil {
			return fmt.Errorf("cannot parse duration string %q: %w", v, err)
		}
		targetValue.SetInt(int64(duration))
		return nil

	case time.Duration:
		targetValue.SetInt(int64(v))
		return nil

	case int64:
		// Numeric values are interpreted as seconds for config-friendly behavior.
		// Use a time.Duration value or a string such as "10s" for finer control.
		duration = time.Duration(v * int64(time.Second))
		targetValue.SetInt(int64(duration))
		return nil

	case int:
		// Numeric values are interpreted as seconds for config-friendly behavior.
		// Use a time.Duration value or a string such as "10s" for finer control.
		duration = time.Duration(v * int(time.Second))
		targetValue.SetInt(int64(duration))
		return nil

	case float64:
		// Numeric values are interpreted as seconds for config-friendly behavior.
		// Fractional seconds are supported, for example 1.5 means 1.5 seconds.
		duration = time.Duration(v * float64(time.Second))
		targetValue.SetInt(int64(duration))
		return nil

	default:
		return fmt.Errorf("cannot decode %T to time.Duration", inputValue)
	}
}

// decodeBasicType decodes basic scalar types.
func (d *Decoder) decodeBasicType(inputValue any, targetValue reflect.Value, targetType reflect.Type) error {
	switch targetType.Kind() {
	case reflect.String:
		return d.decodeToString(inputValue, targetValue)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return d.decodeToInt(inputValue, targetValue, targetType)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return d.decodeToUint(inputValue, targetValue, targetType)
	case reflect.Float32, reflect.Float64:
		return d.decodeToFloat(inputValue, targetValue, targetType)
	case reflect.Bool:
		return d.decodeToBool(inputValue, targetValue)
	default:
		return fmt.Errorf("unsupported type: %s", targetType.Kind())
	}
}

// decodeToString decodes into a string.
func (d *Decoder) decodeToString(inputValue any, targetValue reflect.Value) error {
	switch v := inputValue.(type) {
	case string:
		targetValue.SetString(v)
	case []byte:
		targetValue.SetString(string(v))
	default:
		targetValue.SetString(fmt.Sprintf("%v", v))
	}
	return nil
}

// decodeToInt decodes into a signed integer.
func (d *Decoder) decodeToInt(inputValue any, targetValue reflect.Value, targetType reflect.Type) error {
	var intValue int64

	switch v := inputValue.(type) {
	case int:
		intValue = int64(v)
	case int8:
		intValue = int64(v)
	case int16:
		intValue = int64(v)
	case int32:
		intValue = int64(v)
	case int64:
		intValue = v
	case uint:
		intValue = int64(v)
	case uint8:
		intValue = int64(v)
	case uint16:
		intValue = int64(v)
	case uint32:
		intValue = int64(v)
	case uint64:
		intValue = int64(v)
	case float32:
		intValue = int64(v)
	case float64:
		intValue = int64(v)
	case string:
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot parse string as int: %w", err)
		}
		intValue = parsed
	case bool:
		if v {
			intValue = 1
		} else {
			intValue = 0
		}
	default:
		return fmt.Errorf("cannot decode %T to int", inputValue)
	}

	// Check numeric bounds.
	switch targetType.Kind() {
	case reflect.Int8:
		if intValue < -128 || intValue > 127 {
			return fmt.Errorf("value %d out of range for int8", intValue)
		}
	case reflect.Int16:
		if intValue < -32768 || intValue > 32767 {
			return fmt.Errorf("value %d out of range for int16", intValue)
		}
	case reflect.Int32:
		if intValue < -2147483648 || intValue > 2147483647 {
			return fmt.Errorf("value %d out of range for int32", intValue)
		}
	}

	targetValue.SetInt(intValue)
	return nil
}

// decodeToUint decodes into an unsigned integer.
func (d *Decoder) decodeToUint(inputValue any, targetValue reflect.Value, targetType reflect.Type) error {
	var uintValue uint64

	switch v := inputValue.(type) {
	case uint:
		uintValue = uint64(v)
	case uint8:
		uintValue = uint64(v)
	case uint16:
		uintValue = uint64(v)
	case uint32:
		uintValue = uint64(v)
	case uint64:
		uintValue = v
	case int:
		if v < 0 {
			return fmt.Errorf("cannot decode negative int to uint")
		}
		uintValue = uint64(v)
	case int8:
		if v < 0 {
			return fmt.Errorf("cannot decode negative int8 to uint")
		}
		uintValue = uint64(v)
	case int16:
		if v < 0 {
			return fmt.Errorf("cannot decode negative int16 to uint")
		}
		uintValue = uint64(v)
	case int32:
		if v < 0 {
			return fmt.Errorf("cannot decode negative int32 to uint")
		}
		uintValue = uint64(v)
	case int64:
		if v < 0 {
			return fmt.Errorf("cannot decode negative int64 to uint")
		}
		uintValue = uint64(v)
	case float32:
		if v < 0 {
			return fmt.Errorf("cannot decode negative float32 to uint")
		}
		uintValue = uint64(v)
	case float64:
		if v < 0 {
			return fmt.Errorf("cannot decode negative float64 to uint")
		}
		uintValue = uint64(v)
	case string:
		parsed, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot parse string as uint: %w", err)
		}
		uintValue = parsed
	case bool:
		if v {
			uintValue = 1
		} else {
			uintValue = 0
		}
	default:
		return fmt.Errorf("cannot decode %T to uint", inputValue)
	}

	// Check numeric bounds.
	switch targetType.Kind() {
	case reflect.Uint8:
		if uintValue > 255 {
			return fmt.Errorf("value %d out of range for uint8", uintValue)
		}
	case reflect.Uint16:
		if uintValue > 65535 {
			return fmt.Errorf("value %d out of range for uint16", uintValue)
		}
	case reflect.Uint32:
		if uintValue > 4294967295 {
			return fmt.Errorf("value %d out of range for uint32", uintValue)
		}
	}

	targetValue.SetUint(uintValue)
	return nil
}

// decodeToFloat decodes into a floating-point value.
func (d *Decoder) decodeToFloat(inputValue any, targetValue reflect.Value, targetType reflect.Type) error {
	var floatValue float64

	switch v := inputValue.(type) {
	case float32:
		floatValue = float64(v)
	case float64:
		floatValue = v
	case int:
		floatValue = float64(v)
	case int8:
		floatValue = float64(v)
	case int16:
		floatValue = float64(v)
	case int32:
		floatValue = float64(v)
	case int64:
		floatValue = float64(v)
	case uint:
		floatValue = float64(v)
	case uint8:
		floatValue = float64(v)
	case uint16:
		floatValue = float64(v)
	case uint32:
		floatValue = float64(v)
	case uint64:
		floatValue = float64(v)
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return fmt.Errorf("cannot parse string as float: %w", err)
		}
		floatValue = parsed
	case bool:
		if v {
			floatValue = 1.0
		} else {
			floatValue = 0.0
		}
	default:
		return fmt.Errorf("cannot decode %T to float", inputValue)
	}

	if targetType.Kind() == reflect.Float32 {
		if floatValue > math.MaxFloat32 || floatValue < -math.MaxFloat32 {
			return fmt.Errorf("value %g out of range for float32", floatValue)
		}
	}

	targetValue.SetFloat(floatValue)
	return nil
}

// decodeToBool decodes into a boolean.
func (d *Decoder) decodeToBool(inputValue any, targetValue reflect.Value) error {
	switch v := inputValue.(type) {
	case bool:
		targetValue.SetBool(v)
	case string:
		parsed, err := strconv.ParseBool(v)
		if err != nil {
			return fmt.Errorf("cannot parse string as bool: %w", err)
		}
		targetValue.SetBool(parsed)
	case int:
		targetValue.SetBool(v != 0)
	case int8:
		targetValue.SetBool(v != 0)
	case int16:
		targetValue.SetBool(v != 0)
	case int32:
		targetValue.SetBool(v != 0)
	case int64:
		targetValue.SetBool(v != 0)
	case uint:
		targetValue.SetBool(v != 0)
	case uint8:
		targetValue.SetBool(v != 0)
	case uint16:
		targetValue.SetBool(v != 0)
	case uint32:
		targetValue.SetBool(v != 0)
	case uint64:
		targetValue.SetBool(v != 0)
	case float32:
		targetValue.SetBool(v != 0)
	case float64:
		targetValue.SetBool(v != 0)
	default:
		return fmt.Errorf("cannot decode %T to bool", inputValue)
	}

	return nil
}

// toSlice converts an arbitrary value into a slice.
func (d *Decoder) toSlice(inputValue any) ([]any, bool) {
	switch v := inputValue.(type) {
	case []any:
		return v, true
	case []string:
		result := make([]any, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result, true
	case []int:
		result := make([]any, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result, true
	case []float64:
		result := make([]any, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result, true
	case []bool:
		result := make([]any, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result, true
	default:
		// Try reflection for other slice-like values.
		value := reflect.ValueOf(inputValue)
		if value.Kind() == reflect.Slice {
			result := make([]any, value.Len())
			for i := 0; i < value.Len(); i++ {
				result[i] = value.Index(i).Interface()
			}
			return result, true
		}
		return nil, false
	}
}

// toMap converts an arbitrary value into map[string]any.
func (d *Decoder) toMap(inputValue any) (map[string]any, bool) {
	if inputValue == nil {
		return nil, false
	}

	switch v := inputValue.(type) {
	case map[string]any:
		return v, true
	default:
		value := reflect.ValueOf(inputValue)
		if value.Kind() != reflect.Map {
			return nil, false
		}

		result := make(map[string]any, value.Len())
		iter := value.MapRange()
		for iter.Next() {
			key := iter.Key()
			if key.Kind() != reflect.String {
				return nil, false
			}
			result[key.String()] = iter.Value().Interface()
		}
		return result, true
	}
}
