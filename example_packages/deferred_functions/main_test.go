package deferredfunctions

import (
	"testing"
	"time"
)

func TestX(t *testing.T) {
	defer func() {
		X()
	}()
	time.Sleep(5 * time.Second)
}
