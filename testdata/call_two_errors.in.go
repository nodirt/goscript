package test

func f(g func() (error, error)) {
	g()
}
