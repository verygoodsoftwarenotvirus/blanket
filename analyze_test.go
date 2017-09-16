package main

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/fatih/set"
	"github.com/stretchr/testify/assert"
)

func TestGetDeclaredNamesFromFile(t *testing.T) {
	t.Parallel()

	simple := func(t *testing.T) {
		in, err := parser.ParseFile(token.NewFileSet(), "example_packages/simple/main.go", nil, parser.AllErrors)
		if err != nil {
			t.Logf("failing because ParseFile returned error: %v", err)
			t.FailNow()
		}

		expectedDeclarations := []string{"A", "B", "C", "wrapper"}
		expected := set.New()
		for _, x := range expectedDeclarations {
			expected.Add(x)
		}

		actual := set.New()

		getDeclaredNamesFromFile(in, actual)

		assert.Equal(t, expected, actual, "expected output did not match actual output")
	}

	methods := func(t *testing.T) {
		in, err := parser.ParseFile(token.NewFileSet(), "example_packages/methods/main.go", nil, parser.AllErrors)
		if err != nil {
			t.Logf("failing because ParseFile returned error: %v", err)
			t.FailNow()
		}

		expectedDeclarations := []string{"Example.A", "Example.B", "Example.C", "wrapper"}
		expected := set.New()
		for _, x := range expectedDeclarations {
			expected.Add(x)
		}

		actual := set.New()
		getDeclaredNamesFromFile(in, actual)

		assert.Equal(t, expected, actual, "expected output did not match actual output")
	}

	subtests := []subtest{
		{
			Message: "simple package",
			Test:    simple,
		},
		{
			Message: "methods",
			Test:    methods,
		},
	}
	runSubtestSuite(t, subtests)
}

func TestGetCalledNames(t *testing.T) {
	simple := func(t *testing.T) {
		in, err := parser.ParseFile(token.NewFileSet(), "example_packages/simple/main_test.go", nil, parser.AllErrors)
		if err != nil {
			t.Logf("failing because ParseFile returned error: %v", err)
			t.FailNow()
		}

		expectedDeclarations := []string{"A", "C", "wrapper"}
		expected := set.New()
		for _, x := range expectedDeclarations {
			expected.Add(x)
		}

		actual := set.New()
		getCalledNames(in, actual)

		assert.Equal(t, expected, actual, "expected output did not match actual output")
	}

	methods := func(t *testing.T) {
		in, err := parser.ParseFile(token.NewFileSet(), "example_packages/methods/main_test.go", nil, parser.AllErrors)
		if err != nil {
			t.Logf("failing because ParseFile returned error: %v", err)
			t.FailNow()
		}

		expectedDeclarations := []string{"Example.A", "Example.C", "wrapper"}
		expected := set.New()
		for _, x := range expectedDeclarations {
			expected.Add(x)
		}

		actual := set.New()
		getCalledNames(in, actual)

		assert.Equal(t, expected, actual, "expected output did not match actual output")
	}

	subtests := []subtest{
		{
			Message: "simple package",
			Test:    simple,
		},
		{
			Message: "methods",
			Test:    methods,
		},
	}
	runSubtestSuite(t, subtests)
}
