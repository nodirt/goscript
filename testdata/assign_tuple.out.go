package test

func f(g func() (int, error)) {
	var err0 error
	var a int
	a, err0 = g()
	if err0 != nil {
		panic(err0)
	}
	println(a)
}
