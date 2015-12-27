package test

func f(g func() error) {
	_ = g()
	{
		var err0 error
		println(err0)
	}
}
