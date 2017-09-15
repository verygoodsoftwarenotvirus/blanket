package simple

func A() string {
	return "A"
}

func B() string {
	return "B"
}

func C() string {
	return "C"
}

func wrapper() {
	A()
	B()
	C()
}
