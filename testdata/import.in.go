package test

import "os"

func f() {
	os.RemoveAll("/dev/null")
}
