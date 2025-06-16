// Copyright 2019 Torben Schinke. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jsonptr

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// A Ptr specifies a specific value within a JSON document.
// See https://tools.ietf.org/html/rfc6901 for the specification.
type Ptr = string

// A PtrToken is a single element or token of a Ptr
type PtrToken = string

// Eval takes the json pointer and applies it to the given json object or array.
// Returns an error if the json pointer cannot be resolved.
func Eval(objOrArr ObjOrArr, ptr Ptr) (Value, error) {
	if len(ptr) == 0 {
		// the whole document selector
		return objOrArr, nil
	}

	if !strings.HasPrefix(ptr, "/") {
		return nil, fmt.Errorf("invalid json pointer: %s", ptr)
	}

	tokens := strings.Split(ptr, "/")[1:] // ignore the first empty token
	var root Value
	root = objOrArr
	for tIdx, token := range tokens {
		token = Unescape(token)

		if root == nil {
			return nil, fmt.Errorf("key '%s' not found:\n%s", token, evalMsg(tIdx, tokens, nil))
		}
		switch t := root.(type) {
		case Obj:
			if val, ok := t[token]; ok {
				root = val
			} else {
				root = nil
				return nil, fmt.Errorf("key '%s' not found:\n%s", token, evalMsg(tIdx, tokens, keysAsSlice(t)))
			}

		case *Arr:
			idx, err := strconv.ParseInt(token, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("expected integer index:\n%s", evalMsgArr(tIdx, tokens, 0, t.Len()))
			}
			if idx < 0 || int(idx) >= t.Len() {
				return nil, fmt.Errorf("index out of bounds:\n%s", evalMsgArr(tIdx, tokens, 0, t.Len()))
			}

			root = t.Get(int(idx))
		default:
			return nil, fmt.Errorf("key '%s' not addressable:\n%s", token, evalMsg(tIdx, tokens, nil))
		}
	}

	return root, nil
}

func MustEval(objOrArr ObjOrArr, ptr Ptr) Value {
	v, err := Eval(objOrArr, ptr)
	if err != nil {
		panic(err)
	}

	return v
}

func keysAsSlice[T any](m map[string]T) []string {
	res := make([]string, len(m))[0:0]
	for k := range m {
		res = append(res, k)
	}
	sort.Strings(res)
	return res
}

func baseMsg(failedAt int, tokens []PtrToken) *strings.Builder {
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
func evalMsg(failedAt int, tokens []PtrToken, keysInContext []string) string {
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
func evalMsgArr(failedAt int, tokens []PtrToken, min, max int) string {
	sb := baseMsg(failedAt, tokens)
	sb.WriteString(fmt.Sprintf(" index must be in [%d...%d[", min, max))
	return sb.String()
}

// Escape takes any string and returns a token.
// ~ becomes ~0 and / becomes ~1
func Escape(str string) PtrToken {
	tmp := strings.Replace(str, "~", "~0", -1)
	return strings.Replace(tmp, "/", "~1", -1)
}

// Unescape takes a token and returns the original string.
// ~0 becomes ~ and ~1 becomes /
func Unescape(str PtrToken) string {
	tmp := strings.Replace(str, "~1", "/", -1)
	return strings.Replace(tmp, "~0", "~", -1)
}
