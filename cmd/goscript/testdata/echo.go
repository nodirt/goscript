package main

import (
	"os"
	"strings"
)

func main() {
	println(strings.Join(os.Args[1:], " "))
}
