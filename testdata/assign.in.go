package test

func f(g func() error) {
	_ = g()
}
