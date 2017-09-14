package tabledriven

import (
	"testing"
	"time"
)

func TestX(t *testing.T) {
	testCases := [3]struct {
		runAt time.Time
	}{}

	for range testCases {
		X()
	}
}
