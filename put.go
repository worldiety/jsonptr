package jsonptr

import (
	"fmt"
	"strconv"
	"strings"
)

// Put tries to resolve the given pointer to put the given value. If ptr addresses an array, an index of -1 will
// just append the value. Otherwise, the array is filled up with nil values. Note that any key which looks like
// an integer is considered to be an array index and not an object key.
func Put(objOrArr ObjOrArr, ptr Ptr, value Value) error {
	if len(ptr) == 0 {
		// the whole document selector
		return fmt.Errorf("cannot put on empty JSON pointer")
	}

	if !strings.HasPrefix(ptr, "/") {
		return fmt.Errorf("invalid json pointer: %s", ptr)
	}

	tokens := strings.Split(ptr, "/")[1:] // ignore the first empty token
	var root Value
	root = objOrArr
	for tIdx, token := range tokens {
		token = Unescape(token)
		leaf := tIdx == len(tokens)-1
		nextMayBeArrayIdx := false
		if tIdx <= len(tokens)-2 {
			if _, err := strconv.Atoi(tokens[tIdx+1]); err == nil {
				nextMayBeArrayIdx = true
			}
		}

		if root == nil {
			return fmt.Errorf("key '%s' not found:\n%s", token, evalMsg(tIdx, tokens, nil))
		}

		switch v := root.(type) {
		case Primitive:
			return fmt.Errorf("value '%s' is a primitive:\n%s", token, evalMsg(tIdx, tokens, nil))
		case *Obj:
			if leaf {
				v.Put(token, value)
				return nil
			} else {
				// TODO to be correct we would need a recursive look ahead to decide if next value is object or array index
				if x, ok := v.Get(token); ok {
					root = x
				} else {
					if nextMayBeArrayIdx {
						root = &Arr{}
					} else {
						root = &Obj{}
					}

					v.Put(token, root)
				}

			}
		case *Arr:
			idx, err := strconv.ParseInt(token, 10, 64)
			if err != nil {
				return fmt.Errorf("expected integer index:\n%s", evalMsgArr(tIdx, tokens, 0, v.Len()))
			}

			if !leaf {
				if idx == -1 {
					if nextMayBeArrayIdx {
						root = &Arr{}
					} else {
						root = &Obj{}
					}

					v.Append(root)
				} else {
					if idx < 0 {
						return fmt.Errorf("index out of bounds:\n%s", evalMsgArr(tIdx, tokens, 0, v.Len()))
					} else {
						for range int(idx+1) - v.Len() {
							v.Append(Null{})
						}

						if nextMayBeArrayIdx {
							root = &Arr{}
						} else {
							root = &Obj{}
						}

						v.SetAt(int(idx), root)
					}
				}

				continue

			}

			if idx < 0 {
				v.Append(value)
			} else {
				for range int(idx+1) - v.Len() {
					v.Append(Null{})
				}

				v.SetAt(int(idx), value)
			}

			return nil
		default:
			return fmt.Errorf("unsupported node type:\n%s", evalMsg(tIdx, tokens, nil))
		}
	}

	return fmt.Errorf("cannot put on empty JSON pointer")
}
