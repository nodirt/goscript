package test

func f(g func() error) {
	var err0 error
	err0 = g()
	if err0 != nil {
		panic(err0)
	}
}
