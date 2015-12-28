# goscript

goscript pre-processes .go files to panic on ignored non-nil errors, 
compiles and runs them.

## Example

script.go:

```go
package main

import "os"

func main() {
    os.RemoveAll("/dev/null")
}
```

Running it with goscript will cause a panic and a non-zero exit code.

    goscript run script.go

## Why?

Go is so nice that you may want to write all scripts in Go, the ones that you would write in bash,
but in scripts you may want to exit as soon as an error happens.
goscript rewrites go code to panic on any non-nil errors that were not handled explicitly.

## Error handling

```go
var ignoredError error

func main() {
  // goscript rewrites this call to panic on error
  file, _ := os.Create("file.txt")
  
  // goscript does not rewrite this call because the error is assigned to a variable.
  _, err := file.WriteString("hello")
  if err != nil {
    panic(err)
  }
  
  // this error is intentionally ignored
  ignoredError = file.Close()
}
```

## Installation

    go install github.com/nodirt/goscript/cmd/goscript
