package test

func f(g func() (int, error)) {
	var err0 error
	_, err0 = g()
	if err0 != nil {
		panic(err0)
	}
}
