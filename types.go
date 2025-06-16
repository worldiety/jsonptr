package jsonptr

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
)

type Value interface {
	value()
	String() string
	Bool() bool
	Float64() float64
}

type Primitive interface {
	primitive()
}

type ObjOrArr interface {
	objOrArr()
	Value
}

type Null struct {
}

func (n Null) UnmarshalJSON(bytes []byte) error {
	return nil
}

func (n Null) MarshalJSON() ([]byte, error) {
	return []byte("null"), nil
}

func (n Null) String() string {
	return "null"
}

func (n Null) Bool() bool {
	return false
}

func (n Null) Float64() float64 {
	return 0
}

func (Null) value() {}

func (Null) primitive() {}

type Number float64

func (Number) value() {}

func (Number) primitive() {}

func (n Number) String() string {
	return strconv.FormatFloat(float64(n), 'f', -1, 64)
}

func (n Number) Bool() bool {
	return n == 0
}

func (n Number) Float64() float64 {
	return float64(n)
}

type Obj map[string]Value

func (Obj) value() {}

func (Obj) objOrArr() {}

func (o Obj) String() string {
	buf, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return fmt.Sprintf("%#v", o)
	}

	return string(buf)
}

func (o Obj) Bool() bool {
	return o != nil
}

func (o Obj) Float64() float64 {
	return 0
}

func (o Obj) UnmarshalJSON(bytes []byte) error {
	var tmp map[string]any
	if err := json.Unmarshal(bytes, &tmp); err != nil {
		return err
	}

	clear(o)
	for k, v := range tmp {
		o[k] = ValueOf(v)
	}

	return nil
}

type Arr struct {
	slice []Value
}

func NewArr(slice ...Value) *Arr {
	return &Arr{slice}
}

func (a *Arr) Len() int {
	if a == nil {
		return 0
	}

	return len(a.slice)
}

func (a *Arr) String() string {
	if a == nil {
		return "[]null"
	}

	buf, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return fmt.Sprintf("%#v", a)
	}

	return string(buf)
}

func (a *Arr) Bool() bool {
	return a != nil
}

func (a *Arr) Float64() float64 {
	return 0
}

func (a *Arr) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.slice)
}

func (a *Arr) UnmarshalJSON(bytes []byte) error {
	var tmp []any
	if err := json.Unmarshal(bytes, &tmp); err != nil {
		return err
	}

	a.slice = a.slice[:0]
	a.slice = make([]Value, 0, len(tmp))
	for _, v := range tmp {
		a.slice = append(a.slice, ValueOf(v))
	}

	return nil
}

func (a *Arr) objOrArr() {}

func (*Arr) value() {}

func (a *Arr) Append(v ...Value) *Arr {
	a.slice = append(a.slice, v...)
	return a
}

func (a *Arr) SetAt(idx int, v Value) {
	a.slice[idx] = v
}

func (a *Arr) Get(idx int) Value {
	return a.slice[idx]
}

type Bool bool

func (Bool) value() {}

func (Bool) primitive() {}

func (n Bool) String() string {
	return strconv.FormatBool(bool(n))
}

func (n Bool) Bool() bool {
	return bool(n)
}

func (n Bool) Float64() float64 {
	if n {
		return 1
	} else {
		return 0
	}
}

type String string

func (String) value() {}

func (String) primitive() {}

func (n String) String() string {
	return string(n)
}

func (n String) Bool() bool {
	v, _ := strconv.ParseBool(string(n))
	return v
}

func (n String) Float64() float64 {
	v, _ := strconv.ParseFloat(string(n), 64)
	return v
}

// ValueOf deep copies any basic go value type into the according json type. Unsupported values are mapped to Null.
// Any Value can be marshaled directly to JSON.
func ValueOf(from any) Value {
	switch t := from.(type) {
	case Value:
		return t
	case map[string]any:
		obj := make(Obj)
		for k, v := range t {
			obj[k] = ValueOf(v)
		}
		return obj
	case []any:
		arr := &Arr{}
		for _, v := range t {
			arr.Append(ValueOf(v))
		}

		return arr
	case bool:
		return Bool(t)
	case string:
		return String(t)
	case float64:
		return Number(t)
	case int:
		return Number(t)
	case int64:
		return Number(t)
	case nil:
		return Null{}
	default:
		slog.Error("unknown type: %T", t)
		return Null{}
	}
}
