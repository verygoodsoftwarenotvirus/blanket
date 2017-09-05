package methods

import (
	"testing"
)

func TestA(t *testing.T) {
	e := &Example{}
	e.A()
}

func TestC(t *testing.T) {
	e := &Example{}
	e.C()
}

func TestOuter(t *testing.T) {
	e := &Example{}
	outer(e)
}
