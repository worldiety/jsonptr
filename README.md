# jsonptr [![Travis-CI](https://travis-ci.com/worldiety/jsonptr.svg?branch=master)](https://travis-ci.com/worldiety/jsonptr) [![Go Report Card](https://goreportcard.com/badge/github.com/worldiety/jsonptr)](https://goreportcard.com/report/github.com/worldiety/jsonptr) [![GoDoc](https://godoc.org/github.com/worldiety/jsonptr?status.svg)](http://godoc.org/github.com/worldiety/jsonptr) [![Sourcegraph](https://sourcegraph.com/github.com/worldiety/jsonptr/-/badge.svg)](https://sourcegraph.com/github.com/worldiety/jsonptr?badge) [![Coverage](http://gocover.io/_badge/github.com/worldiety/jsonptr)](http://gocover.io/github.com/worldiety/jsonptr) 
An implementation of rfc6901 (json pointer) in Go (golang)

## example

Example json
```json
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
```

Api usage
```go
// get an array entry
num,err := jsonptr.Evaluate(obj, "/details/nested/list/2") // returns string value 3

// get an array entry and parse it
i, err := jsonptr.AsInt(obj,"/details/nested/list/0") // returns int value 1
```

In case of evaluation errors while resolving the json pointer, you will get nice error messages.
```bash
     key 'abc' not found:
        /abc/asd
         ^~~~~~~ available keys: (details|msg)
     key 'asd' not found:
        /details/asd
                 ^~~~ available keys: (author|flag|id|name|nested|nice|num)
     key 'x' not found:
        /details/nested/x
                        ^~ available keys: (even|id|list)
     index out of bounds:
        /details/nested/list/4
                             ^~ index must be in [0...3[
     expected integer index:
        /details/nested/list/a4
                             ^~~ index must be in [0...3[
     key 'a' not found:
        /a/b/c/d
         ^~~~~ object is nil
     key 'x' not found:
        /details/nice/x
                      ^~ object is nil
```