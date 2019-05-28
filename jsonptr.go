// Copyright 2019 Torben Schinke. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jsonptr

import (
	"fmt"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// A JSONPointer specifies a specific value within a JSON document.
// See https://tools.ietf.org/html/rfc6901 for the specification.
type JSONPointer = string

// A JSONPointerToken is a single element or token of a JSONPointer
type JSONPointerToken = string

// Evaluate takes the json pointer and applies it to the given json object or array.
// Returns an error if the json pointer cannot be resolved.
func Evaluate(objOrArr interface{}, ptr JSONPointer) (interface{}, error) {
	if len(ptr) == 0 {
		// the whole document selector
		return objOrArr, nil
	}
	if !strings.HasPrefix(ptr, "/") {
		return nil, fmt.Errorf("invalid json pointer: %s", ptr)
	}

	tokens := strings.Split(ptr, "/")[1:] // ignore the first empty token
	var root interface{}
	root = objOrArr
	for tIdx, token := range tokens {
		token = Unescape(token)

	typeSwitch:
		if root == nil {
			return nil, fmt.Errorf("key '%s' not found:\n%s", token, evalMsg(tIdx, tokens, nil))
		}
		switch t := root.(type) {
		case map[string]interface{}:
			if val, ok := t[token]; ok {
				root = val
			} else {
				root = nil
				return nil, fmt.Errorf("key '%s' not found:\n%s", token, evalMsg(tIdx, tokens, keysAsSlice(t)))
			}

		case *map[string]interface{}:
			root = *t
			goto typeSwitch
		case []interface{}:
			idx, err := strconv.ParseInt(token, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("expected integer index:\n%s", evalMsgArr(tIdx, tokens, 0, len(t)))
			}
			if idx < 0 || int(idx) >= len(t) {
				return nil, fmt.Errorf("index out of bounds:\n%s", evalMsgArr(tIdx, tokens, 0, len(t)))
			}
			root = t[idx]

		case *[]interface{}:
			root = *t
			goto typeSwitch
		}
	}
	return root, nil
}

func keysAsSlice(m map[string]interface{}) []string {
	res := make([]string, len(m))[0:0]
	for k := range m {
		res = append(res, k)
	}
	sort.Strings(res)
	return res
}

func baseMsg(failedAt int, tokens []JSONPointerToken) *strings.Builder {
	tmp := &strings.Builder{}
	sb := &strings.Builder{}
	for i, t := range tokens {
		sb.WriteString("/")
		sb.WriteString(t)
		if i < failedAt {
			for s := 0; s < len(t); s++ {
				tmp.WriteString(" ")
			}
			tmp.WriteString(" ") // for the added /
		} else {
			if i == failedAt {
				tmp.WriteString(" ^")
			}
			for s := 0; s < len(t); s++ {
				tmp.WriteString("~")
			}
		}
	}
	sb.WriteString("\n")
	sb.WriteString(tmp.String())
	return sb
}

// evalMsg generates an over-engineered error message
func evalMsg(failedAt int, tokens []JSONPointerToken, keysInContext []string) string {
	sb := baseMsg(failedAt, tokens)

	if len(keysInContext) > 0 {
		sb.WriteString(" available keys: ")
		sb.WriteString("(")
		for kid, key := range keysInContext {
			sb.WriteString(key)
			if kid < len(keysInContext)-1 {
				sb.WriteString("|")
			}
		}
		sb.WriteString(")")
	}

	if keysInContext == nil {
		sb.WriteString(" object is nil")
	}

	return sb.String()
}

// evalMsgArr generates an over-engineered error message
func evalMsgArr(failedAt int, tokens []JSONPointerToken, min, max int) string {
	sb := baseMsg(failedAt, tokens)
	sb.WriteString(fmt.Sprintf(" index must be in [%d...%d[", min, max))
	return sb.String()
}

// Escape takes any string and returns a token.
// ~ becomes ~0 and / becomes ~1
func Escape(str string) JSONPointerToken {
	tmp := strings.Replace(str, "~", "~0", -1)
	return strings.Replace(tmp, "/", "~1", -1)
}

// Unescape takes a token and returns the original string.
// ~0 becomes ~ and ~1 becomes /
func Unescape(str JSONPointerToken) string {
	tmp := strings.Replace(str, "~1", "/", -1)
	return strings.Replace(tmp, "~0", "~", -1)
}

// AsString takes a JSONPointer and tries to interpret the result as a string.
// The following rules are applied:
//  * numbers are converted to the according string representation
//  * booleans are converted to true|false
//  * null is converted to the empty string
//  * a non-resolvable value, returns the empty string
//  * arrays and objects are converted into a json string
//  * String() methods are used, if available
//  * Anything else is converted using sprintf and %v directive
//  * only evaluation errors are returned as errors.
func AsString(objOrArr interface{}, ptr JSONPointer) (string, error) {
	val, err := Evaluate(objOrArr, ptr)
	return toString(val), err
}

// TryString is a go-like "try macro" pattern. Returns the empty string in case of error, otherwise v.
func TryString(v string, err error) string {
	if err != nil {
		return ""
	}
	return v
}

// MustString is a go-like "expect macro" pattern. Panics in case of error, otherwise returns v.
func MustString(v string, err error) string {
	if err != nil {
		panic(err)
	}
	return v
}

// AsFloat64 takes a JSONPointer and tries to interpret the result as a float.
// The following rules are applied:
//  * numbers are converted to a float
//  * booleans are converted to 1|0
//  * null is converted NaN
//  * a non-resolvable value, returns also NaN
//  * arrays and objects are converted to NaN
//  * String() method is invoked, if available, and output parsed. Returns NaN if not parsable.
//  * Anything else is converted using sprintf and %v directive and tried to be parsed. Returns NaN if not parsable.
//  * only evaluation errors are returned as errors.
func AsFloat64(objOrArr interface{}, ptr JSONPointer) (float64, error) {
	val, err := Evaluate(objOrArr, ptr)
	return toFloat64(val), err
}

// TryFloat64 is a go-like "try macro" pattern. Returns NaN in case of error, otherwise v.
func TryFloat64(v float64, err error) float64 {
	if err != nil {
		return math.NaN()
	}
	return v
}

// MustFloat64 is a go-like "expect macro" pattern. Panics in case of error, otherwise returns v.
func MustFloat64(v float64, err error) float64 {
	if err != nil {
		panic(err)
	}
	return v
}

// AsFloat32 takes a JSONPointer and tries to interpret the result as a float.
// The following rules are applied:
//  * numbers are converted to a float
//  * booleans are converted to 1|0
//  * null is converted NaN
//  * a non-resolvable value, returns also NaN
//  * arrays and objects are converted to NaN
//  * String() method is invoked, if available, and output parsed. Returns NaN if not parsable.
//  * Anything else is converted using sprintf and %v directive and tried to be parsed. Returns NaN if not parsable.
//  * only evaluation errors are returned as errors.
func AsFloat32(objOrArr interface{}, ptr JSONPointer) (float32, error) {
	val, err := Evaluate(objOrArr, ptr)
	return float32(toFloat64(val)), err
}

// TryFloat32 is a go-like "try macro" pattern. Returns +Inf in case of error, otherwise v.
func TryFloat32(v float32, err error) float32 {
	if err != nil {
		return float32(math.Inf(+1))
	}
	return v
}

// MustFloat32 is a go-like "expect macro" pattern. Panics in case of error, otherwise returns v.
func MustFloat32(v float32, err error) float32 {
	if err != nil {
		panic(err)
	}
	return v
}

// AsInt takes a JSONPointer and tries to interpret the result as an int using the rules of AsFloat.
// NaN is treated as 0.
//  * only evaluation errors are returned as errors.
func AsInt(objOrArr interface{}, ptr JSONPointer) (int, error) {
	f, err := AsFloat64(objOrArr, ptr)
	if err != nil {
		return 0, err
	}
	if math.IsNaN(f) {
		return 0, nil
	}
	return int(f), nil
}

// TryInt is a go-like "try macro" pattern. Returns 0 in case of error, otherwise v.
func TryInt(v int, err error) int {
	if err != nil {
		return 0
	}
	return v
}

// MustInt is a go-like "expect macro" pattern. Panics in case of error, otherwise returns v.
func MustInt(v int, err error) int {
	if err != nil {
		panic(err)
	}
	return v
}

// AsInt32 takes a JSONPointer and tries to interpret the result as an int using the rules of AsFloat.
// NaN is treated as 0.
//  * only evaluation errors are returned as errors.
func AsInt32(objOrArr interface{}, ptr JSONPointer) (int32, error) {
	f, err := AsFloat64(objOrArr, ptr)
	if err != nil {
		return 0, err
	}
	if math.IsNaN(f) {
		return 0, nil
	}
	return int32(f), nil
}

// TryInt32 is a go-like "try macro" pattern. Returns 0 in case of error, otherwise v.
func TryInt32(v int32, err error) int32 {
	if err != nil {
		return 0
	}
	return v
}

// MustInt32 is a go-like "expect macro" pattern. Panics in case of error, otherwise returns v.
func MustInt32(v int32, err error) int32 {
	if err != nil {
		panic(err)
	}
	return v
}

// AsInt64 takes a JSONPointer and tries to interpret the result as an int using the rules of AsFloat.
// NaN is treated as 0.
//  * only evaluation errors are returned as errors.
func AsInt64(objOrArr interface{}, ptr JSONPointer) (int64, error) {
	f, err := AsFloat64(objOrArr, ptr)
	if err != nil {
		return 0, err
	}
	if math.IsNaN(f) {
		return 0, nil
	}
	return int64(f), nil
}

// TryInt32 is a go-like "try macro" pattern. Returns 0 in case of error, otherwise v.
func TryInt64(v int64, err error) int64 {
	if err != nil {
		return 0
	}
	return v
}

// MustInt32 is a go-like "expect macro" pattern. Panics in case of error, otherwise returns v.
func MustInt64(v int64, err error) int64 {
	if err != nil {
		panic(err)
	}
	return v
}

// AsBool takes a JSONPointer and tries to interpret the result as a bool.
// Anything else is interpreted as false.
//  * only evaluation errors are returned as errors.
func AsBool(objOrArr interface{}, ptr JSONPointer) (bool, error) {
	val, err := Evaluate(objOrArr, ptr)
	if err != nil {
		return false, err
	}

	b, err := toBool(val)
	if err != nil {
		return false, nil
	}
	return b, nil
}

// TryBool is a go-like "try macro" pattern. Returns false in case of error, otherwise v.
func TryBool(v bool, err error) bool {
	if err != nil {
		return false
	}
	return v
}

// MustBool is a go-like "expect macro" pattern. Panics in case of error, otherwise returns v.
func MustBool(v bool, err error) bool {
	if err != nil {
		panic(err)
	}
	return v
}

// AsArray evaluates the JSONPointer and unwraps either []interface{} or *[]interface{} slices. Any other slice
// types are unboxed to []interface{}. If value is not a slice type at all, the value is inserted into a one-element
// interface slice.
//  * only evaluation errors are returned as errors.
func AsArray(objOrArr interface{}, ptr JSONPointer) ([]interface{}, error) {
	val, err := Evaluate(objOrArr, ptr)
	if err != nil {
		return nil, err
	}
	switch t := val.(type) {
	case []interface{}:
		return t, nil
	case *[]interface{}:
		return *t, nil
	default:
		s := reflect.ValueOf(t)
		if s.Kind() != reflect.Slice {
			return []interface{}{t}, nil
		}

		ret := make([]interface{}, s.Len())
		for i := 0; i < s.Len(); i++ {
			ret[i] = s.Index(i).Interface()
		}

		return ret, nil
	}
}

// TryArray is a go-like "try macro" pattern. Returns nil in case of error, otherwise v.
func TryArray(v []interface{}, err error) []interface{} {
	if err != nil {
		return nil
	}
	return v
}

// MustArray is a go-like "expect macro" pattern. Panics in case of error, otherwise returns v.
func MustArray(v []interface{}, err error) []interface{} {
	if err != nil {
		panic(err)
	}
	return v
}

// AsArray evaluates the JSONPointer and unwraps map[string]interface{}. Any other slice
// types are unboxed to []interface{}. If value is not a map type at all, the value is inserted into a one-element
// map with the key value.
//  * only evaluation errors are returned as errors.
func AsObject(objOrArr interface{}, ptr JSONPointer) (map[string]interface{}, error) {
	val, err := Evaluate(objOrArr, ptr)
	if err != nil {
		return nil, err
	}
	switch t := val.(type) {
	case map[string]interface{}:
		return t, nil
	default:
		s := reflect.ValueOf(t)
		if s.Kind() != reflect.Map {
			return map[string]interface{}{"value": t}, nil
		}

		ret := make(map[string]interface{})
		for _, _ = range s.MapKeys() {
			//ret[k] = s.
			panic("fix me")
		}

		return ret, nil
	}
}

// TryObject is a go-like "try macro" pattern. Returns empty map in case of error, otherwise v.
func TryObject(v map[string]interface{}, err error) map[string]interface{} {
	if err != nil {
		return nil
	}
	return v
}

// MustObject is a go-like "expect macro" pattern. Panics in case of error, otherwise returns v.
func MustObject(v map[string]interface{}, err error) map[string]interface{} {
	if err != nil {
		panic(err)
	}
	return v
}

// AsFloat64Array evaluates the JSONPointer and tries to interpret any slice value as float (see AsFloat) for rules
//  * only evaluation errors are returned as errors.
func AsFloat64Array(objOrArr interface{}, ptr JSONPointer) ([]float64, error) {
	slice, err := AsArray(objOrArr, ptr)
	if err != nil {
		return nil, err
	}
	res := make([]float64, len(slice))
	for i, v := range slice {
		res[i] = toFloat64(v)
	}
	return res, nil
}

// AsStringArray evaluates the JSONPointer and tries to interpret any slice value as string (see AsString) for rules
//  * only evaluation errors are returned as errors.
func AsStringArray(objOrArr interface{}, ptr JSONPointer) ([]string, error) {
	slice, err := AsArray(objOrArr, ptr)
	if err != nil {
		return nil, err
	}
	res := make([]string, len(slice))
	for i, v := range slice {
		res[i] = toString(v)
	}
	return res, nil
}

// TryStringArray is a go-like "try macro" pattern. Returns nil in case of error, otherwise v.
func TryStringArray(v []string, err error) []string {
	if err != nil {
		return nil
	}
	return v
}

// MustStringArray is a go-like "expect macro" pattern. Panics in case of error, otherwise returns v.
func MustStringArray(v []string, err error) []string {
	if err != nil {
		panic(err)
	}
	return v
}

// AsIntArray evaluates the JSONPointer and uses the AsFloatArray conversion rules but truncates to int.
// JSON does not support integers at all, just floats anyway. NaN values are treated as 0.
//  * only evaluation errors are returned as errors.
func AsIntArray(objOrArr interface{}, ptr JSONPointer) ([]int, error) {
	slice, err := AsArray(objOrArr, ptr)
	if err != nil {
		return nil, err
	}
	res := make([]int, len(slice))
	for i, v := range slice {
		f := toFloat64(v)
		if math.IsNaN(f) {
			res[i] = 0
		} else {
			res[i] = int(toFloat64(v))
		}

	}
	return res, nil
}
