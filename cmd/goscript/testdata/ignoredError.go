package main

import "os"

var ignore error

func main() {
	ignore = os.Remove("/dev/null")
}