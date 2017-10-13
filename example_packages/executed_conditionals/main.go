package executed_conditionals

func a(condition bool) string {
	if condition {
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

func wrapper(condition bool) {
	a(condition)
	b()
	c()
}
