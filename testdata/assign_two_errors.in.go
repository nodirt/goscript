package test

func f(g func() (error, error)) {
	_, _ = g()
}
