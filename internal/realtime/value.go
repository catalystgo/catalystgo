package realtime

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// Value describes the value of a config
type Value interface {
	IsNil() bool
	IsEqual(Value) bool

	Bool() bool
	MaybeBool() (bool, error)

	Int() int
	MaybeInt() (int, error)

	Int64() int64
	MaybeInt64() (int64, error)

	Uint() uint
	MaybeUint() (uint, error)

	Uint64() uint64
	MaybeUint64() (uint64, error)

	Float64() float64
	MaybeFloat64() (float64, error)

	Duration() time.Duration
	MaybeDuration() (time.Duration, error)

	String() string
	MaybeString() (string, error)

	ParsedJSON() map[string]interface{}

	MaybeParsedJSON() (map[string]interface{}, error)
}

type value struct {
	values []interface{}
}

// NewValue creates a new instance of value
// It properly parses passed interface, appends all elements if it's a slice or an array
func NewValue(vv interface{}) value {
	if vv == nil {
		return value{values: nil}
	}
	rv := reflect.ValueOf(vv)
	if rv.Kind() == reflect.Array || rv.Kind() == reflect.Slice {
		if length := rv.Len(); length > 0 {
			values := []interface{}{}
			for i := 0; i < length; i++ {
				elem := rv.Index(i).Interface()
				if val, ok := elem.(value); ok {
					values = append(values, val.values...)
					continue
				}
				values = append(values, elem)
			}
			return value{values: values}
		}
		return value{values: nil}
	}
	return value{values: []interface{}{vv}}
}

// IsNil reports if the underlying value is nil
func (v value) IsNil() bool {
	return len(v.values) == 0 || len(v.values) == 1 && v.values[0] == nil
}

// IsEqual reports if the underlying value is equal to the passed one
func (v value) IsEqual(vv Value) bool {
	return v.String() == vv.String()
}

// Bool returns the underlying value as bool
func (v value) Bool() bool {
	val, _ := v.MaybeBool()
	return val
}

// MaybeBool returns the underlying value as bool with error (if any)
func (v value) MaybeBool() (bool, error) {
	raw := v.String()
	val, err := strconv.ParseBool(raw)

	if err != nil {
		return false, fmt.Errorf("value cannot be parsed as bool: %v", err)
	}
	return val, nil
}

// Int returns the underlying value as int
func (v value) Int() int {
	val, _ := v.MaybeInt()
	return val
}

// MaybeInt returns the underlying value as int with error (if any)
func (v value) MaybeInt() (int, error) {
	raw := v.String()
	val, err := strconv.Atoi(raw)

	if err != nil {
		return 0, fmt.Errorf("value cannot be parsed as int: %v", err)
	}
	return val, nil
}

// Int64 returns the underlying value as int64
func (v value) Int64() int64 {
	val, _ := v.MaybeInt64()
	return val
}

// MaybeInt64 returns the underlying value as int64 with error (if any)
func (v value) MaybeInt64() (int64, error) {
	raw := v.String()
	val, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("value cannot be parsed as int64: %v", err)
	}
	return val, nil
}

// Uint returns the underlying value as uint
func (v value) Uint() uint {
	val, _ := v.MaybeUint()
	return val
}

// MaybeUint returns the underlying value as uint with error (if any)
func (v value) MaybeUint() (uint, error) {
	raw := v.String()
	val, err := strconv.ParseUint(raw, 10, 32)

	if err != nil {
		return 0, fmt.Errorf("value cannot be parsed as uint: %v", err)
	}
	return uint(val), nil
}

// Uint64 returns the underlying value as uint64
func (v value) Uint64() uint64 {
	val, _ := v.MaybeUint64()
	return val
}

// MaybeUint64 returns the underlying value as uint64 with error (if any)
func (v value) MaybeUint64() (uint64, error) {
	raw := v.String()
	val, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("value cannot be parsed as uint64: %v", err)
	}
	return val, nil
}

// Float64 returns the underlying value as float64
func (v value) Float64() float64 {
	val, _ := v.MaybeFloat64()
	return val
}

// MaybeFloat64 returns the underlying value as float64 with error (if any)
func (v value) MaybeFloat64() (float64, error) {
	raw := v.String()
	val, err := strconv.ParseFloat(raw, 64)

	if err != nil {
		return 0, fmt.Errorf("value cannot be parsed as float64: %v", err)
	}
	return val, nil
}

// Duration returns the underlying value as time.Duration
func (v value) Duration() time.Duration {
	val, _ := v.MaybeDuration()
	return val
}

// MaybeDuration returns the underlying value as time.Duration with error (if any)
func (v value) MaybeDuration() (time.Duration, error) {
	raw := v.String()
	val, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("value cannot be parsed as time.Duration: %v", err)
	}
	return val, nil
}

// String returns the underlying value as string
func (v value) String() string {
	raw, _ := v.MaybeString()
	return raw
}

// MaybeString returns the underlying value as string with error (though it's always nil)
func (v value) MaybeString() (string, error) {
	if v.IsNil() {
		return "", nil
	}
	if len(v.values) == 1 {
		return fmt.Sprintf("%v", v.values[0]), nil
	}

	return fmt.Sprintf("%v", v.values), nil
}
func (v value) GoString() string {
	return v.String()
}
func (v value) ParsedJSON() map[string]interface{} {
	val, _ := v.MaybeParsedJSON()
	return val
}

// ParsedJSON returns the underlying  map and error if there is.
func (v value) MaybeParsedJSON() (map[string]interface{}, error) {
	raw := v.String()
	m := make(map[string]interface{})
	err := json.Unmarshal([]byte(raw), &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// ExtractValuesFromValue allows to extract internal value slice from Value
func ExtractValuesFromValue(v Value) []Value {
	val, ok := v.(value)
	if !ok {
		return []Value{v}
	}
	result := make([]Value, len(val.values))
	for i := range val.values {
		result[i] = NewValue(val.values[i])
	}
	return result
}

var _ Value = &NilValue{}

// NilValue implements zero value for Value.
type NilValue struct {
	err error
}

// NewNilValue creates a new instance of NilValue with the given error.

func NewNilValue(err error) NilValue {
	return NilValue{err}
}

// IsNil always return true.

func (v NilValue) IsNil() bool { return true }

// IsEqual always return false.

func (v NilValue) IsEqual(Value) bool { return false }

// Bool always return false.

func (v NilValue) Bool() bool { return false }

// MaybeBool returns false and error if it exists.

func (v NilValue) MaybeBool() (bool, error) { return false, v.err }

// Int always return 0.

func (v NilValue) Int() int { return 0 }

// MaybeInt returns 0 and error if it exists.

func (v NilValue) MaybeInt() (int, error) { return 0, v.err }

// Int64 always return 0.

func (v NilValue) Int64() int64 { return 0 }

// MaybeInt64 returns 0 and error if it exists.

func (v NilValue) MaybeInt64() (int64, error) { return 0, v.err }

// Uint always return 0.

func (v NilValue) Uint() uint { return 0 }

// MaybeUint returns 0 and error if it exists.

func (v NilValue) MaybeUint() (uint, error) { return 0, v.err }

// Uint64 always return 0.

func (v NilValue) Uint64() uint64 { return 0 }

// MaybeUint64 returns 0 and error if it exists.

func (v NilValue) MaybeUint64() (uint64, error) { return 0, v.err }

// Float64 always return 0.

func (v NilValue) Float64() float64 { return 0 }

// MaybeFloat64 returns 0 and error if it exists.

func (v NilValue) MaybeFloat64() (float64, error) { return 0, v.err }

// Duration always return 0.

func (v NilValue) Duration() time.Duration { return 0 }

// MaybeDuration returns 0 and error if it exists.

func (v NilValue) MaybeDuration() (time.Duration, error) { return 0, v.err }

// String always return empty string.

func (v NilValue) String() string { return "" }

// MaybeString returns empty string and error if it exists.

func (v NilValue) MaybeString() (string, error) { return "", v.err }

// ParsedJSON returns nil map.

func (v NilValue) ParsedJSON() map[string]interface{} { return nil }

// MaybeParsedJSON returns nil map and error if it exists.

func (v NilValue) MaybeParsedJSON() (map[string]interface{}, error) { return nil, v.err }

// Error returns error, if it exists.

func (v NilValue) Error() error {
	return v.err
}
