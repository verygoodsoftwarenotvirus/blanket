package methods

import (
	"testing"
)

var (
	a example
)

func helperGenerator(t *testing.T) (*example, error) {
	t.Helper()
	return &example{}, nil
}

func TestA(t *testing.T) {
	a.A()
}

func TestB(t *testing.T) {
	b := example{}
	b.B()
}

func TestC(t *testing.T) {
	c := &example{}
	c.D()

}

func TestD(t *testing.T) {
	d, _ := helperGenerator(t)
	d.C()
}

func TestE(t *testing.T) {
	var e example
	e.E()
}

func TestWrapper(t *testing.T) {
	wrapper(&example{})
}
