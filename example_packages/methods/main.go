package methods

type Example struct{}

func (e *Example) A() string {
	return "A"
}

func (e *Example) B() string {
	return "B"
}

func (e Example) C() string {
	return "C"
}

func (e Example) D() string {
	return "D"
}

func (e Example) E() string {
	return "E"
}

func (e Example) F() string {
	return "F"
}

func wrapper(e *Example) {
	e.A()
	e.B()
	e.C()
	e.D()
	e.E()
	e.F()
}
