package test

func f(g func() error) {
	var err1 error
	err1 = g()
	if err1 != nil {
		panic(err1)
	}
	{
		var err0 error
		println(err0)
	}
}
