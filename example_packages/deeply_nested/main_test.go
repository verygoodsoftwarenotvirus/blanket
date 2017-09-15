package deeply_nested

import (
	"testing"
)

func TestX(t *testing.T) {
	for range [10]struct{}{} {
		for range [5]struct{}{} {
			for range [1]struct{}{} {
				X()
			}
		}
	}
}
