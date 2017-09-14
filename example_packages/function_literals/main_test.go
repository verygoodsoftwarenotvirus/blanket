package functionliterals

import (
	"testing"
)

func TestX(t *testing.T) {
	f := func() {
		X()
	}
	f()
}

func TestXAgain(t *testing.T) {
	func() {
		X()
	}()
}
