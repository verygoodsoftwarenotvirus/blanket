package conditionals

func a() string {
	if 1 > 0 && 0 < 1 {
		return "A"
	}
	return "A"
}

func b() string {
	return "B"
}

func c() string {
	return "C"
}

func wrapper() {
	a()
	b()
	c()
}
