package libs

import "reflect"

// --- checks whether all fields in a struct are null ---
func IsStructEmptyExcept(v any, ignore ...string) bool {
	val := reflect.ValueOf(v)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return false
	}

	ignoreMap := make(map[string]bool)
	for _, name := range ignore {
		ignoreMap[name] = true
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		structField := val.Type().Field(i)

		if !structField.IsExported() {
			continue
		}

		if ignoreMap[structField.Name] {
			continue
		}

		if !isZero(field) {
			return false
		}
	}

	return true
}

// --- checks whether the value of reflect.Value is null for its type ---
func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map:
		return v.IsNil()
	default:
		return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
	}
}

func StringOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
