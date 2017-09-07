package main

import (
	"go/parser"
	"go/token"
	"os"
	"testing"

	"github.com/fatih/set"
	"github.com/stretchr/testify/assert"
)

func TestGetDeclaredNamesWithSimpleFunctions(t *testing.T) {
	t.Parallel()

	in, err := parser.ParseFile(token.NewFileSet(), "example_packages/simple/main.go", nil, parser.AllErrors)
	if err != nil {
		t.Logf("failing because ParseFile returned error: %v", err)
		t.FailNow()
	}

	expectedDeclarations := []string{"A", "B", "C", "outer"}
	expected := set.New()
	for _, x := range expectedDeclarations {
		expected.Add(x)
	}

	actual := set.New()

	getDeclaredNames(in, actual)

	assert.Equal(t, expected, actual, "expected output did not match actual output")
}

func TestGetDeclaredNamesWithStructMethods(t *testing.T) {
	t.Parallel()

	in, err := parser.ParseFile(token.NewFileSet(), "example_packages/methods/main.go", nil, parser.AllErrors)
	if err != nil {
		t.Logf("failing because ParseFile returned error: %v", err)
		t.FailNow()
	}

	expectedDeclarations := []string{"Example.A", "Example.B", "Example.C", "outer"}
	expected := set.New()
	for _, x := range expectedDeclarations {
		expected.Add(x)
	}

	actual := set.New()

	getDeclaredNames(in, actual)

	assert.Equal(t, expected, actual, "expected output did not match actual output")
}

func TestSimplePackage(t *testing.T) {
	t.Parallel()
	originalArgs := os.Args
	os.Args = []string{
		originalArgs[0],
		"--package=github.com/verygoodsoftwarenotvirus/veneer/example_packages/simple",
	}

	main()
	os.Args = originalArgs
}
