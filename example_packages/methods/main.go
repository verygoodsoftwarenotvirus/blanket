package methods

type example struct{}

func (e *example) A() string {
	return "A"
}

func (e *example) B() string {
	return "B"
}

func (e example) C() string {
	return "C"
}

func (e example) D() string {
	return "D"
}

func (e example) E() string {
	return "E"
}

func (e example) F() string {
	return "F"
}

func wrapper(e *example) {
	e.A()
	e.B()
	e.C()
	e.D()
	e.E()
	e.F()
}
