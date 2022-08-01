# struct_mapper [![GoDoc](https://godoc.org/github.com/gvassili/struct_mapper?status.svg)](https://godoc.org/github.com/gvassili/struct_mapper)

A golang struct to struct mapping package using reflection.
> :warning: **Work in progress**
>
todo list
- [ ] Go documentation
- [X] Mapping to a structure
- [ ] Mapping from a structure
- [ ] Write unit test
- [ ] Investigate thread safety on `GetDecoder` method
- [ ] Handle custom interface to transform the value like the `json.Unmarshaler` interface

## Installation
``` sh
go get github.com/gvassili/strcut_mapper
```

## Example
``` go
package main

import (
    "fmt"
    "github.com/gvassili/struc_mapper"
)

type Source struct {
    Field1 string `map:"Field4"`
    Field2 int `map:"Field3"`
}

type Destination struct {
    Field3 int
    Field4 string
}

func main() {
    mapper := struct_mapper.New()
    src := Source{"abc", 123}
    var dst Destination
    decoder, err := mapper.GetDecoder(&src, &dst)
    if err != nil {
        panic(fmt.Errorf("struct mapper decoder decode src into dst: %w", err))
    }
    if err := decoder.Decode(&src, &dst); err != nil {
        panic(fmt.Errorf("struct mapper decoder decode src into dst: %w", err))
    }
    fmt.Printf("src: %+v\ndst: %+v\n", src, dst)
}
```
output:
```
src: {Field1:abc Field2:123}
dst: {Field3:123 Field4:abc}
```

It's possible to have complexes writing path by setting an array on the tag `path` (refer to the [tag_parser package](https://github.com/gvassili/tag_parser) for parsing rules). The tree structure of the source doesn't matter. Also, Pointer type destination field are allocated if `nil`.
``` go
type Source struct {
	Field1 string `map:"path=SubPath2 Field4"`
	Field2 *int   `map:"path=SubPath1 Field3"`
}

type Destination struct {
	SubPath1 struct {
		Field3 int
	}
	SubPath2 *struct {
		Field4 string
	}
}
```

It's also possible to add custom handler when transforming the value and/or the type is required.
``` go
package main

import (
	"fmt"
	"time"
    "github.com/gvassili/struc_mapper"
)

type Source struct {
	GoTimeTs time.Time `map:"UnixTs,unix-time"`
}

type Destination struct {
	UnixTs int64
}

func UnixTimeHandler(v any) any {
	return v.(time.Time).Unix()
}

func main() {
	mapper := struct_mapper.New()
	ts, err := time.Parse(time.RFC3339, "1970-01-01T00:02:03Z")
	src := Source{ts}
	var dst Destination
	mapper.SetHandle("unix-time", UnixTimeHandler)
	decoder, err := mapper.GetDecoder(&src, &dst)
	if err != nil {
		panic(fmt.Errorf("struct mapper decoder decode src into dst: %w", err))
	}
	if err := decoder.Decode(&src, &dst); err != nil {
		panic(fmt.Errorf("struct mapper decoder decode src into dst: %w", err))
	}
	fmt.Printf("src: %+v\ndst: %+v\n", src, dst)
}
```

output
```
src: {GoTimeTs:1970-01-01 00:02:03 +0000 UTC}
dst: {UnixTs:123}
```

