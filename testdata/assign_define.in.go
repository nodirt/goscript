package test

func f(g func() (int, error)) {
	a, _ := g()
	println(a)
}
