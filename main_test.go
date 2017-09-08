package main

import (
	"go/parser"
	"go/token"
	"log"
	"os"
	"testing"

	"github.com/bouk/monkey"
	"github.com/fatih/set"
	"github.com/stretchr/testify/assert"
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
	testPassed := true
	for _, test := range tests {
		if !testPassed {
			t.FailNow()
		}
		testPassed = t.Run(test.Message, test.Test)
	}
}

////////////////////////////////////////////////////////
//                                                    //
//                    Actual Tests                    //
//                                                    //
////////////////////////////////////////////////////////

func TestGetDeclaredNames(t *testing.T) {
	t.Parallel()

	simple := func(t *testing.T) {
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

	methods := func(t *testing.T) {
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
		in, err := parser.ParseFile(token.NewFileSet(), "example_packages/methods/main_test.go", nil, parser.AllErrors)
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

func TestMain(t *testing.T) {
	originalArgs := os.Args

	optimal := func(t *testing.T) {
		os.Args = []string{
			originalArgs[0],
			"--package=github.com/verygoodsoftwarenotvirus/tarp/example_packages/simple",
		}

		main()
		os.Args = originalArgs
	}

	nonexistentPackage := func(t *testing.T) {
		os.Args = []string{
			originalArgs[0],
			"--package=github.com/nosuchrealusername/absolutelynosuchpackage",
			"--fail-on-extras",
		}

		defer func() {
			if r := recover(); r != nil {
				// recovered from our monkey patched log.Fatalf
				assert.True(t, true)
			}
		}()

		var fatalfCalled bool
		monkey.Patch(log.Fatalf, func(string, ...interface{}) {
			fatalfCalled = true
			panic("log.Fatalf")
		})

		main()
		assert.True(t, fatalfCalled, "main should call log.Fatalf() when --fail-on-extras is passed in and extras are found")

		os.Args = originalArgs
		monkey.Unpatch(log.Fatalf)
	}

	invalidCode := func(t *testing.T) {
		os.Args = []string{
			originalArgs[0],
			"--package=github.com/verygoodsoftwarenotvirus/tarp/example_packages/invalid",
			"--fail-on-extras",
		}

		defer func() {
			if r := recover(); r != nil {
				// recovered from our monkey patched log.Fatal
				assert.True(t, true)
			}
		}()

		var fatalCalled bool
		monkey.Patch(log.Fatal, func(...interface{}) {
			fatalCalled = true
			panic("log.Fatal")
		})

		main()

		assert.True(t, fatalCalled, "main should call log.Fatal() when --fail-on-extras is passed in and extras are found")

		os.Args = originalArgs
		monkey.Unpatch(log.Fatal)
	}

	invalidArguments := func(t *testing.T) {
		os.Args = []string{originalArgs[0]}
		defer func() {
			if r := recover(); r != nil {
				// recovered from our monkey patched log.Fatal
				assert.True(t, true)
			}
		}()

		var fatalCalled bool
		monkey.Patch(log.Fatal, func(...interface{}) {
			fatalCalled = true
			panic("log.Fatal")
		})

		main()
		assert.True(t, fatalCalled, "main should call log.Fatal when --fail-on-extras is passed in and extras are found")
		os.Args = originalArgs
		monkey.Unpatch(log.Fatal)
	}

	failsWhenInstructed := func(t *testing.T) {
		os.Args = []string{
			originalArgs[0],
			"--package=github.com/verygoodsoftwarenotvirus/tarp/example_packages/simple",
			"--fail-on-extras",
		}

		var fatalCalled bool
		monkey.Patch(log.Fatal, func(...interface{}) {
			fatalCalled = true
		})

		main()
		assert.True(t, fatalCalled, "main should call log.Fatal() when --fail-on-extras is passed in and extras are found")
		os.Args = originalArgs
		monkey.Unpatch(log.Fatal)
	}

	subtests := []subtest{
		{
			Message: "optimal",
			Test:    optimal,
		},
		{
			Message: "nonexistent package",
			Test:    nonexistentPackage,
		},
		{
			Message: "invalid code",
			Test:    invalidCode,
		},
		{
			Message: "invalid args",
			Test:    invalidArguments,
		},
		{
			Message: "fails with --fail-on-extras",
			Test:    failsWhenInstructed,
		},
	}
	runSubtestSuite(t, subtests)
}
