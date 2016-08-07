go-tail-win
===========
[![GoDoc](https://godoc.org/github.com/brunoqc/go-tail-win?status.svg)](https://godoc.org/github.com/brunoqc/go-tail-win)
[![Build status](https://ci.appveyor.com/api/projects/status/7wwyhxu523it1nu7?svg=true)](https://ci.appveyor.com/project/brunoqc/go-tail-win)

A Go package that behaves like tail.

##Sample code##
```go
package main

import (
	"fmt"

	"github.com/brunoqc/go-tail-win"
)

func main() {
    t, err := tail.TailFile(filePath)
	if err != nil {
		panic(err)
	}

	for line := range t.Lines {
        fmt.Printf("> %q\n", line)
	}
}
```
