package mapstruct

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// 自定义错误类型
var (
	ErrArrayLengthMismatch = errors.New("array length mismatch")
)

// Decoder 提供 map[string]any 到 struct 的解码功能
type Decoder struct {
	// TagName 指定用于字段映射的标签名
	// 默认为空字符串（使用字段名作为key）
	// 可选值: "mapstruct", "json", 或其他自定义标签名, 默认使用 "yaml" 标签
	TagName string
	// StrictMode 严格模式，如果字段类型不匹配则返回错误
	StrictMode bool
	// TimeLayout 时间解析格式，默认为 RFC3339
	TimeLayout string
}

// New 创建一个新的解码器
func New() *Decoder {
	return &Decoder{
		TagName:    "yaml", // 默认使用字段名作为key
		StrictMode: false,
		TimeLayout: time.RFC3339,
	}
}

// WithTagName 设置标签名
func (d *Decoder) WithTagName(tagName string) *Decoder {
	d.TagName = tagName
	return d
}

// WithStrictMode 设置严格模式
func (d *Decoder) WithStrictMode(strict bool) *Decoder {
	d.StrictMode = strict
	return d
}

// WithTimeLayout 设置时间格式
func (d *Decoder) WithTimeLayout(layout string) *Decoder {
	d.TimeLayout = layout
	return d
}

// Decode 将 map[string]any 解码为指定的结构体
func (d *Decoder) Decode(input map[string]any, target interface{}) error {
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer")
	}

	targetValue = targetValue.Elem()
	targetType := targetValue.Type()

	if targetType.Kind() != reflect.Struct {
		return fmt.Errorf("target must be a struct")
	}

	return d.decodeStruct(input, targetValue, targetType)
}

// decodeStruct 解码结构体
func (d *Decoder) decodeStruct(input map[string]any, targetValue reflect.Value, targetType reflect.Type) error {
	for i := 0; i < targetType.NumField(); i++ {
		field := targetType.Field(i)
		fieldValue := targetValue.Field(i)

		// 跳过不可设置的字段
		if !fieldValue.CanSet() {
			continue
		}

		// 根据字段类型选择解码方法
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

// decodeEmbeddedField 解码内嵌字段
func (d *Decoder) decodeEmbeddedField(input map[string]any, fieldValue reflect.Value, field reflect.StructField) error {
	if err := d.decodeStruct(input, fieldValue, field.Type); err != nil {
		return d.handleDecodeError(err, field.Name)
	}
	return nil
}

// decodeNormalField 解码普通字段
func (d *Decoder) decodeNormalField(input map[string]any, fieldValue reflect.Value, field reflect.StructField) error {
	// 获取字段的映射名称
	fieldName := d.getFieldName(field)
	if fieldName == "" {
		return nil
	}

	// 从输入中获取值
	inputValue, exists := input[fieldName]
	if !exists {
		return nil
	}

	// 解码字段值
	if err := d.decodeField(inputValue, fieldValue); err != nil {
		return d.handleDecodeError(err, fieldName)
	}

	return nil
}

// handleDecodeError 处理解码错误，根据错误类型和模式决定是否返回错误
func (d *Decoder) handleDecodeError(err error, fieldName string) error {
	// 数组长度不匹配始终返回错误
	if errors.Is(err, ErrArrayLengthMismatch) {
		return fmt.Errorf("failed to decode field %s: %w", fieldName, err)
	}

	// 严格模式下返回所有错误
	if d.StrictMode {
		return fmt.Errorf("failed to decode field %s: %w", fieldName, err)
	}

	// 非严格模式下忽略错误
	return nil
}

// getFieldName 获取字段的映射名称
func (d *Decoder) getFieldName(field reflect.StructField) string {
	// 如果没有指定TagName，直接使用字段名
	if d.TagName == "" {
		return field.Name
	}

	// 尝试获取标签
	tag := field.Tag.Get(d.TagName)
	if tag == "" {
		return field.Name
	}

	// 标签值为 "-" 表示忽略该字段
	if tag == "-" {
		return ""
	}

	return tag
}

// decodeField 解码单个字段
func (d *Decoder) decodeField(inputValue any, targetValue reflect.Value) error {
	targetType := targetValue.Type()

	// 如果输入值为 nil，跳过
	if inputValue == nil {
		return nil
	}

	// 如果类型完全匹配，直接赋值
	if reflect.TypeOf(inputValue) == targetType {
		targetValue.Set(reflect.ValueOf(inputValue))
		return nil
	}

	// 处理指针类型
	if targetType.Kind() == reflect.Ptr {
		return d.decodeToPointer(inputValue, targetValue, targetType)
	}

	// 处理切片类型
	if targetType.Kind() == reflect.Slice {
		return d.decodeToSlice(inputValue, targetValue, targetType)
	}

	// 处理数组类型
	if targetType.Kind() == reflect.Array {
		return d.decodeToArray(inputValue, targetValue, targetType)
	}

	// 处理结构体类型
	if targetType.Kind() == reflect.Struct {
		// 特殊处理 time.Time
		if targetType.String() == "time.Time" {
			return d.decodeToTime(inputValue, targetValue)
		}
		return d.decodeToStruct(inputValue, targetValue, targetType)
	}

	// 处理基本类型
	return d.decodeBasicType(inputValue, targetValue, targetType)
}

// decodeToPointer 解码到指针类型
func (d *Decoder) decodeToPointer(inputValue any, targetValue reflect.Value, targetType reflect.Type) error {
	elemType := targetType.Elem()
	elemValue := reflect.New(elemType).Elem()

	if err := d.decodeField(inputValue, elemValue); err != nil {
		return err
	}

	targetValue.Set(elemValue.Addr())
	return nil
}

// decodeToSlice 解码到切片类型
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

// decodeToArray 解码到数组类型
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

// decodeToStruct 解码到结构体类型
func (d *Decoder) decodeToStruct(inputValue any, targetValue reflect.Value, targetType reflect.Type) error {
	inputMap, ok := inputValue.(map[string]any)
	if !ok {
		return fmt.Errorf("cannot decode %T to struct", inputValue)
	}

	return d.decodeStruct(inputMap, targetValue, targetType)
}

// decodeToTime 解码到 time.Time 类型
func (d *Decoder) decodeToTime(inputValue any, targetValue reflect.Value) error {
	var t time.Time
	var err error

	switch v := inputValue.(type) {
	case string:
		// 尝试多种时间格式
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
		// Unix timestamp (可能包含小数秒)
		sec := int64(v)
		nsec := int64((v - float64(sec)) * 1e9)
		t = time.Unix(sec, nsec)
		targetValue.Set(reflect.ValueOf(t))
		return nil

	default:
		return fmt.Errorf("cannot decode %T to time.Time", inputValue)
	}
}

// decodeBasicType 解码基本类型
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

// decodeToString 解码为字符串
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

// decodeToInt 解码为整数
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

	// 检查范围
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

// decodeToUint 解码为无符号整数
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

	// 检查范围
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

// decodeToFloat 解码为浮点数
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
		targetValue.SetFloat(floatValue)
	} else {
		targetValue.SetFloat(floatValue)
	}

	return nil
}

// decodeToBool 解码为布尔值
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

// toSlice 将任意值转换为切片
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
		// 尝试使用反射处理其他切片类型
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
