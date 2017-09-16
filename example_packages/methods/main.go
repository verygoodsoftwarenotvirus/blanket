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

func wrapper(e *Example) {
	e.A()
	e.B()
	e.C()
}
