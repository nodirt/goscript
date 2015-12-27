package test

func f(g func() (int, error)) {
	var a int
	a, _ = g()
	println(a)
}
