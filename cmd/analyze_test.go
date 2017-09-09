package tarp

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/fatih/set"
	"github.com/stretchr/testify/assert"
	"fmt"
)

////////////////////////////////////////////////////////
//                                                    //
//               Test Helper Functions                //
//                                                    //
////////////////////////////////////////////////////////

type subtest struct {
	Message string
	Test    func(t *testing.T)
}

func runSubtestSuite(t *testing.T, tests []subtest) {
	t.Helper()
	for _, test := range tests {
		t.Run(test.Message, test.Test)
	}
}

func buildExamplePackagePath(t *testing.T, pkg string) string {
	t.Helper()
	return buildPackagePath(fmt.Sprintf("github.com/verygoodsoftwarenotvirus/tarp/example_packages/%s", pkg))
}

////////////////////////////////////////////////////////
//                                                    //
//                    Actual Tests                    //
//                                                    //
////////////////////////////////////////////////////////

func TestGetDeclaredNames(t *testing.T) {
	t.Parallel()

	simple := func(t *testing.T) {
		in, err := parser.ParseFile(token.NewFileSet(), buildExamplePackagePath(t, "simple/main.go"), nil, parser.AllErrors)
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

	methods := func(t *testing.T) {
		in, err := parser.ParseFile(token.NewFileSet(), buildExamplePackagePath(t, "methods/main.go"), nil, parser.AllErrors)
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
	t.Parallel()

	simple := func(t *testing.T) {
		in, err := parser.ParseFile(token.NewFileSet(), buildExamplePackagePath(t, "simple/main_test.go"), nil, parser.AllErrors)
		if err != nil {
			t.Logf("failing because ParseFile returned error: %v", err)
			t.FailNow()
		}

		expectedDeclarations := []string{"A", "C", "outer"}
		expected := set.New()
		for _, x := range expectedDeclarations {
			expected.Add(x)
		}

		actual := set.New()
		getCalledNames(in, actual)

		assert.Equal(t, expected, actual, "expected output did not match actual output")
	}

	methods := func(t *testing.T) {
		in, err := parser.ParseFile(token.NewFileSet(), buildExamplePackagePath(t, "methods/main_test.go"), nil, parser.AllErrors)
		if err != nil {
			t.Logf("failing because ParseFile returned error: %v", err)
			t.FailNow()
		}

		expectedDeclarations := []string{".Parallel", "Example.A", "Example.C", "outer"}
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
