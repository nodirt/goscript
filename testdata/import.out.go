package test

import "os"

func f() {
	var err0 error
	err0 = os.RemoveAll("/dev/null")
	if err0 != nil {
		panic(err0)
	}
}
