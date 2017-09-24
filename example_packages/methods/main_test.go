package methods

import (
	"testing"
)

var (
	a Example
)

func helperGenerator(t *testing.T) (*Example, error) {
	t.Helper()
	return &Example{}, nil
}

func TestA(t *testing.T) {
	a.A()
}

func TestB(t *testing.T) {
	b := Example{}
	b.B()
}

func TestC(t *testing.T) {
	c := &Example{}
	c.D()

}

func TestD(t *testing.T) {
	d, _ := helperGenerator(t)
	d.C()
}

func TestE(t *testing.T) {
	var e Example
	e.E()
}

func TestWrapper(t *testing.T) {
	wrapper(&Example{})
}
