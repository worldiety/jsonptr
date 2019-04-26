// Copyright 2019 Torben Schinke. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jsonptr

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
)

// toBool converts anything to a bool
func toBool(v interface{}) (bool, error) {
	switch t := v.(type) {
	case bool:
		return t, nil
	case int:
		if t == 0 {
			return false, nil
		} else if t == 1 {
			return true, nil
		}
	}
	return strconv.ParseBool(toString(v))
}

// toString converts anything to a string
func toString(any interface{}) string {
	if any == nil {
		return ""
	}
	switch t := any.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatFloat(t, 'g', -1, 64)
	case int64:
		return strconv.FormatInt(t, 10)
	case bool:
		return strconv.FormatBool(t)
	case fmt.Stringer:
		return t.String()
	case []interface{}:
		data, err := json.Marshal(t)
		if err != nil {
			//ups, unable to json marshal that thingy
			return fmt.Sprintf("%v", any)
		}
		return string(data)
	case map[string]interface{}:
		data, err := json.Marshal(t)
		if err != nil {
			//ups, unable to json marshal that thingy
			return fmt.Sprintf("%v", any)
		}
		return string(data)

	case *[]interface{}:
		return toString(*t)
	}
	return fmt.Sprintf("%v", any)
}

// toFloat64 converts anything to a float
func toFloat64(any interface{}) float64 {
	if any == nil {
		return math.NaN()
	}
	switch t := any.(type) {

	case float64:
		return t
	case int64:
		return float64(t)
	case uint64:
		return float64(t)
	case int32:
		return float64(t)
	case uint32:
		return float64(t)
	case int16:
		return float64(t)
	case uint16:
		return float64(t)
	case uint8:
		return float64(t)
	case int8:
		return float64(t)
	case bool:
		if t {
			return 1
		}
		return 0

	case string:
		return parseStrAsFloat(t)
	case fmt.Stringer:
		return parseStrAsFloat(t.String())
	default:
		return parseStrAsFloat(fmt.Sprintf("%v", any))

	}

}

func parseStrAsFloat(str string) float64 {
	if f, err := strconv.ParseFloat(str, 64); err == nil {
		return f
	}
	return math.NaN()
}
