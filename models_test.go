package main

import (
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlanketDetailsMethods(t *testing.T) {
	arbitraryInstance := blanketDetails{
		blanketFunc{
			Filename: "a",
			Name:     "One",
		},
		blanketFunc{
			Filename: "b",
			Name:     "Two",
			DeclPos: token.Position{
				Line: 1,
			},
		},
		blanketFunc{
			Filename: "b",
			Name:     "Three",
			DeclPos: token.Position{
				Line: 2,
			},
		},
	}

	t.Run(".Len()", func(_t *testing.T) {
		assert.Equal(t, 3, arbitraryInstance.Len(), ".Len() should return the length of blanketDetails")
	})

	t.Run(".Less()", func(_t *testing.T) {
		assert.True(t, arbitraryInstance.Less(0, 1), ".Less(i, j) should return the correct response")
		assert.False(t, arbitraryInstance.Less(1, 0), ".Less(i, j) should return the correct response")
		assert.True(t, arbitraryInstance.Less(1, 2), ".Less(i, j) should return the correct response")
	})

	t.Run(".Swap()", func(_t *testing.T) {
		expected := blanketDetails{
			blanketFunc{
				Filename: "b",
				Name:     "Two",
				DeclPos: token.Position{
					Line: 1,
				},
			},
			blanketFunc{
				Filename: "a",
				Name:     "One",
			},
			blanketFunc{
				Filename: "b",
				Name:     "Three",
				DeclPos: token.Position{
					Line: 2,
				},
			},
		}
		arbitraryInstance.Swap(0, 1)
		assert.Equal(t, expected, arbitraryInstance, ".Swap(i, j) should swap the location of two values")
	})
}
