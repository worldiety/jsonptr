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
	obj := &Obj{}
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

	res, err := Eval(obj, "")
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(obj, res) {
		t.Fatal("expected the same")
	}

	res, err = Eval(obj, "/foo")
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(res, NewArr(String("bar"), String("baz"))) {
		t.Fatal("expected the same")
	}

}

func TestPrintErr(t *testing.T) {
	obj := &Obj{}
	err := json.Unmarshal([]byte(testJSON2), &obj)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Eval(obj, "/abc/asd")
	t.Log(err)

	_, err = Eval(obj, "/details/asd")
	t.Log(err)

	_, err = Eval(obj, "/details/nested/x")
	t.Log(err)

	_, err = Eval(obj, "/details/nested/list/4")
	t.Log(err)

	_, err = Eval(obj, "/details/nested/list/a4")
	t.Log(err)

	_, err = Eval(nil, "/a/b/c/d")
	t.Log(err)

	_, err = Eval(obj, "/details/nice/x")
	t.Log(err)
}

func TestHelper(t *testing.T) {
	obj := &Obj{}
	err := json.Unmarshal([]byte(testJSON2), &obj)
	if err != nil {
		t.Fatal(err)
	}

	if v := MustEval(obj, "/details/id").Float64(); v != 123 {
		t.Fatal("unexpected", v)
	}

	if v := MustEval(obj, "/details/num").Float64(); v != 3.14 {
		t.Fatal("unexpected", v)
	}

	if v := MustEval(obj, "/details/num").Bool(); v != false {
		t.Fatal("unexpected", v)
	}

	if v := MustEval(obj, "/details/flag").Bool(); v != true {
		t.Fatal("unexpected", v)
	}

	if v := MustEval(obj, "/details/num").String(); v != "3.14" {
		t.Fatal("unexpected", v)
	}

	if v := MustEval(obj, "/details/flag").String(); v != "true" {
		t.Fatal("unexpected", v)
	}

	if v := MustEval(obj, "/details/flag").Float64(); v != 1 {
		t.Fatal("unexpected", v)
	}

	arr := MustEval(obj, "/details/nested/list").(*Arr)
	if arr.Len() != 3 {
		t.Fatal("unexpected len", arr.Len())
	}
	if arr.Get(1).Float64() != 2 {
		t.Fatal("unexpected", arr.Get(1))
	}

	if v := MustEval(obj, "/details/nested").(*Obj); v.Len() != 3 {
		t.Fatal("unexpected", v)
	}

}

func expectStr(t *testing.T, json ObjOrArr, ptr Ptr, val string) {
	t.Helper()
	v, err := Eval(json, ptr)
	if err != nil {
		t.Fatal(err)
	}

	if str, ok := v.(String); ok {
		if str.String() != val {
			t.Fatal("expected", val, "but got", str)
		}
	} else {
		t.Fatal("unexpected", v)
	}
}

func expectFloat64(t *testing.T, json ObjOrArr, ptr Ptr, val float64) {
	t.Helper()
	v, err := Eval(json, ptr)
	if err != nil {
		t.Fatal(err)
	}

	if str, ok := v.(Number); ok {
		if str.Float64() != val {
			t.Fatal("expected", val, "but got", str)
		}
	} else {
		t.Fatal("unexpected", v)
	}
}
