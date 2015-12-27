package main

import (
	"os"
	"fmt"
)

func main() {
	fmt.Println("Hello World")
	os.RemoveAll("/dev/null")
}
