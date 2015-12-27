package test

func f(g func() (error, error)) {
	var err0, err1 error
	err0, err1 = g()
	if err0 != nil {
		panic(err0)
	}
	if err1 != nil {
		panic(err1)
	}
}
