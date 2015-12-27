package test

func f(g func() (int, error)) {
	g()
}
