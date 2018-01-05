package executedconditionals

import (
	"testing"
)

func TestB(t *testing.T) {
	b()
}

func TestC(t *testing.T) {
	c()
}

func TestWrapper(t *testing.T) {
	wrapper(true)
}
