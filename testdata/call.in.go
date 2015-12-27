package test

func f(g func() error) {
	g()
}
