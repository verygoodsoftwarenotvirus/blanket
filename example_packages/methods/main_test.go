package methods

import (
	"testing"
)

var (
	f Example
	g Example
)

func helperGenerator(t *testing.T) (*Example, error) {
	t.Helper()
	return &Example{}, nil
}

func TestA(t *testing.T) {
	t.Parallel()
	var e Example
	e.A()
}

func TestC(t *testing.T) {
	x, err := helperGenerator(t)
	if err != nil {
		t.FailNow()
	}
	x.C()
}

//func TestCAgain(t *testing.T) {
//	f.C()
//}

func TestWrapper(t *testing.T) {
	e := &Example{}
	wrapper(e)
}
