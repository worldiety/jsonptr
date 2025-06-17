package jsonptr

import (
	"encoding/json"
	"testing"
)

func TestPut(t *testing.T) {
	obj := &Obj{}

	if err := Put(obj, "/a/b/c", String("hello")); err != nil {
		t.Fatal(err)
	}

	if err := Put(obj, "/a/b/array/-1", Number(5)); err != nil {
		t.Fatal(err)
	}

	if err := Put(obj, "/a/b/array/2", Number(42)); err != nil {
		t.Fatal(err)
	}

	if err := Put(obj, "/a/b/array/-1/hello/-1", Bool(true)); err != nil {
		t.Fatal(err)
	}

	t.Log(obj.String())

	if v := MustEval(obj, "/a/b/c").String(); v != "hello" {
		t.Fatalf(`expected "hello", got "%s"`, v)
	}

	if v := MustEval(obj, "/a/b/array/0").Float64(); v != 5 {
		t.Fatalf(`expected 5, got %f`, v)
	}

	if v := MustEval(obj, "/a/b/array/1").String(); v != "null" {
		t.Fatalf(`expected "null", got "%s"`, v)
	}

	if v := MustEval(obj, "/a/b/array/2").String(); v != "42" {
		t.Fatalf(`expected "42", got "%s"`, v)
	}

	if v := MustEval(obj, "/a/b/array/3/hello/0").String(); v != "true" {
		t.Fatalf(`expected "42", got "%s"`, v)
	}
}

func TestUnmarshal(t *testing.T) {
	var obj Obj
	if err := json.Unmarshal([]byte(`{"a":1}`), &obj); err != nil {
		t.Fatal(err)
	}

	t.Log(obj.String())

	if v, _ := obj.Get("a"); v.Float64() != 1 {
		t.Fatalf(`expected "1", got %v`, v)
	}

	var arr []Obj
	if err := json.Unmarshal([]byte(`[{"a":1},{"a":2}]`), &arr); err != nil {
		t.Fatal(err)
	}

	if v, _ := arr[1].Get("a"); v.Float64() != 2 {
		t.Fatalf(`expected "2", got %v`, v)
	}

	t.Log(arr)
}
