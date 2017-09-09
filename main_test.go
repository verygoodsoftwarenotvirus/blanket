package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/bouk/monkey"
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
	t.Helper()
	for _, test := range tests {
		t.Run(test.Message, test.Test)
	}
}

func buildExamplePackagePath(t *testing.T, pkg string) string {
	t.Helper()
	gopath := os.Getenv("GOPATH")
	path := fmt.Sprintf("github.com/verygoodsoftwarenotvirus/tarp/example_packages/%s", pkg)
	pkgDir := strings.Join([]string{gopath, "src", path}, "/")
	return pkgDir
}

func createInvalidCodeFile(t *testing.T) {
	/*
		so basically if this file is in the `example_packages`
		folder, I can't run `go test ./...`, which I value more
		than bearing the shame of having this awful function in
		my tests
	*/
	t.Helper()

	err := os.MkdirAll(buildExamplePackagePath(t, "invalid"), os.ModePerm)
	assert.Nil(t, err, "no error should be encountered trying to create the invalid temp folder.")
	if err != nil {
		t.FailNow()
	}

	f, err := os.Create(buildExamplePackagePath(t, "invalid/main.go"))
	assert.Nil(t, err, "no error should be encountered trying to create the invalid temp file.")
	if err != nil {
		t.FailNow()
	}
	fmt.Fprint(f, `
		package invalid

		import (
			"log"


		funk main() {
			return e
		)
	`)
}

func deleteInvalidCodeFile(t *testing.T) {
	t.Helper()
	err := os.RemoveAll(buildExamplePackagePath(t, "invalid"))
	assert.Nil(t, err, "no error should be encountered trying to delete the invalid temp folder.")
	if err != nil {
		t.FailNow()
	}
}

////////////////////////////////////////////////////////
//                                                    //
//                    Actual Tests                    //
//                                                    //
////////////////////////////////////////////////////////

func TestMainFart(t *testing.T) {
	originalArgs := os.Args

	optimal := func(t *testing.T) {
		os.Args = []string{
			originalArgs[0],
			"analyze",
			"--package=github.com/verygoodsoftwarenotvirus/tarp/example_packages/simple",
		}

		main()
		os.Args = originalArgs
	}

	nonexistentPackage := func(t *testing.T) {
		os.Args = []string{
			originalArgs[0],
			"analyze",
			"--package=github.com/nosuchrealusername/absolutelynosuchpackage",
			"--fail-on-finding",
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
		assert.True(t, fatalfCalled, "main should call log.Fatalf() when --fail-on-finding is passed in and extras are found")

		os.Args = originalArgs
		monkey.Unpatch(log.Fatalf)
	}

	invalidCode := func(t *testing.T) {
		createInvalidCodeFile(t)
		os.Args = []string{
			originalArgs[0],
			"analyze",
			"--package=github.com/verygoodsoftwarenotvirus/tarp/example_packages/invalid",
			"--fail-on-finding",
		}

		defer func() {
			if r := recover(); r != nil {
				// recovered from our monkey patched log.Fatal
				assert.True(t, true)
				deleteInvalidCodeFile(t)
			}
		}()

		var fatalCalled bool
		monkey.Patch(log.Fatal, func(...interface{}) {
			fatalCalled = true
			panic("log.Fatal")
		})

		main()

		assert.True(t, fatalCalled, "main should call log.Fatal() when --fail-on-finding is passed in and extras are found")
		os.Args = originalArgs
		monkey.Unpatch(log.Fatal)
	}

	invalidArguments := func(t *testing.T) {
		os.Args = []string{
			originalArgs[0],
			"analyze",
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
		assert.True(t, fatalCalled, "main should call log.Fatal when invalid arguments are passed to analyze")
		os.Args = originalArgs
		monkey.Unpatch(log.Fatal)
	}

	failsWhenInstructed := func(t *testing.T) {
		os.Args = []string{
			originalArgs[0],
			"analyze",
			"--package=github.com/verygoodsoftwarenotvirus/tarp/example_packages/simple",
			"--fail-on-finding",
		}

		var fatalCalled bool
		monkey.Patch(log.Fatal, func(...interface{}) {
			fatalCalled = true
		})

		main()
		assert.True(t, fatalCalled, "main should call log.Fatal() when --fail-on-finding is passed in and extras are found")
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
			Message: "fails with --fail-on-finding",
			Test:    failsWhenInstructed,
		},
	}
	runSubtestSuite(t, subtests)
}
