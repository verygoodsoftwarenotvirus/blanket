package methods

import (
	"testing"
)

var (
	f Example
	g Example
)

func TestA(t *testing.T) {
	var e Example
	e.A()
}

func TestC(t *testing.T) {
	g.C()
}

func TestOuter(t *testing.T) {
	e := &Example{}
	outer(e)
}
