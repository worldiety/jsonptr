// Copyright 2019 Torben Schinke. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jsonptr

import (
	"encoding/json"
	"reflect"
	"testing"
)

const rfcTestJSON = `

   {
      "foo": ["bar", "baz"],
      "": 0,
      "a/b": 1,
      "c%d": 2,
      "e^f": 3,
      "g|h": 4,
      "i\\j": 5,
      "k\"l": 6,
      " ": 7,
      "m~n": 8
   }
`

const testJSON2 = `
{
	"details":{
		"name":"hello",
		"id":123,
		"num":3.14,
		"flag":true,
		"author":"A Name",
		"nested":{
			"even":"more",
			"id":4,
			"list":["1","2","3"]
	
		},
		"nice":null
	},
	"msg":"hello"
}
`

func TestEvaluate(t *testing.T) {
	obj := make(map[string]interface{})
	err := json.Unmarshal([]byte(rfcTestJSON), &obj)
	if err != nil {
		t.Fatal(err)
	}
	expectStr(t, obj, "/foo/0", "bar")
	expectFloat64(t, obj, "/", 0)
	expectFloat64(t, obj, "/a~1b", 1)
	expectFloat64(t, obj, "/c%d", 2)
	expectFloat64(t, obj, "/e^f", 3)
	expectFloat64(t, obj, "/g|h", 4)
	expectFloat64(t, obj, "/i\\j", 5)
	expectFloat64(t, obj, "/k\"l", 6)
	expectFloat64(t, obj, "/ ", 7)
	expectFloat64(t, obj, "/m~0n", 8)

	res, err := Evaluate(obj, "")
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(obj, res) {
		t.Fatal("expected the same")
	}

	res, err = Evaluate(obj, "/foo")
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(res, ([]interface{}{"bar", "baz"})) {
		t.Fatal("expected the same")
	}

}

func TestPrintErr(t *testing.T) {
	obj := make(map[string]interface{})
	err := json.Unmarshal([]byte(testJSON2), &obj)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Evaluate(obj, "/abc/asd")
	t.Log(err)

	_, err = Evaluate(obj, "/details/asd")
	t.Log(err)

	_, err = Evaluate(obj, "/details/nested/x")
	t.Log(err)

	_, err = Evaluate(obj, "/details/nested/list/4")
	t.Log(err)

	_, err = Evaluate(obj, "/details/nested/list/a4")
	t.Log(err)

	_, err = Evaluate(nil, "/a/b/c/d")
	t.Log(err)

	_, err = Evaluate(obj, "/details/nice/x")
	t.Log(err)
}

func TestHelper(t *testing.T) {
	obj := make(map[string]interface{})
	err := json.Unmarshal([]byte(testJSON2), &obj)
	if err != nil {
		t.Fatal(err)
	}

	if v, _ := AsFloat(obj, "/details/id"); v != 123 {
		t.Fatal("unexpected", v)
	}

	if v, _ := AsFloat(obj, "/details/num"); v != 3.14 {
		t.Fatal("unexpected", v)
	}

	if v, _ := AsBool(obj, "/details/num"); v != false {
		t.Fatal("unexpected", v)
	}

	if v, _ := AsBool(obj, "/details/flag"); v != true {
		t.Fatal("unexpected", v)
	}

	if v, _ := AsString(obj, "/details/num"); v != "3.14" {
		t.Fatal("unexpected", v)
	}

	if v, _ := AsString(obj, "/details/flag"); v != "true" {
		t.Fatal("unexpected", v)
	}

	if v, _ := AsInt(obj, "/details/flag"); v != 1 {
		t.Fatal("unexpected", v)
	}

	arr, _ := AsFloatArray(obj, "/details/nested/list")
	if len(arr) != 3 {
		t.Fatal("unexpected len", len(arr))
	}
	if arr[1] != 2 {
		t.Fatal("unexpected", arr[1])
	}

	if v, _ := AsFloatArray(obj, "/details/nested"); len(v) != 1 {
		t.Fatal("unexpected", v)
	}

	if v, _ := AsStringArray(obj, "/details/nested"); len(v) != 1 {
		t.Fatal("unexpected", v)
	}

	if v, _ := AsString(obj, "/details/nested"); len(v) == 0 {
		t.Fatal("unexpected", v)
	}
}

func expectStr(t *testing.T, json interface{}, ptr JSONPointer, val string) {
	t.Helper()
	v, err := Evaluate(json, ptr)
	if err != nil {
		t.Fatal(err)
	}

	if str, ok := v.(string); ok {
		if str != val {
			t.Fatal("expected", val, "but got", str)
		}
	} else {
		t.Fatal("unexpected", v)
	}
}

func expectFloat64(t *testing.T, json interface{}, ptr JSONPointer, val float64) {
	t.Helper()
	v, err := Evaluate(json, ptr)
	if err != nil {
		t.Fatal(err)
	}

	if str, ok := v.(float64); ok {
		if str != val {
			t.Fatal("expected", val, "but got", str)
		}
	} else {
		t.Fatal("unexpected", v)
	}
}
