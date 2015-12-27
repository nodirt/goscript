package test

func f(g func() int) {
	_ = g()
}
