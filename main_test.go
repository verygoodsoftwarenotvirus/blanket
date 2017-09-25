// +build !unit

package main

import (
	"fmt"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"

	"errors"
	"github.com/bouk/monkey"
	"github.com/fatih/set"
	"github.com/stretchr/testify/assert"
)

////////////////////////////////////////////////////////
//                                                    //
//               Test Helper Functions                //
//                                                    //
////////////////////////////////////////////////////////

func buildExamplePackagePath(t *testing.T, packageName string, abs bool) string {
	t.Helper()
	gopath := os.Getenv("GOPATH")
	if abs {
		return strings.Join([]string{gopath, "src", "github.com", "verygoodsoftwarenotvirus", "tarp", "example_packages", packageName}, "/")
	}
	return strings.Join([]string{"github.com", "verygoodsoftwarenotvirus", "tarp", "example_packages", packageName}, "/")
}

////////////////////////////////////////////////////////
//                                                    //
//                   Actual Tests                     //
//                                                    //
////////////////////////////////////////////////////////

func init() {
	log.SetOutput(ioutil.Discard)

	monkey.Patch(log.Fatalf, func(string, ...interface{}) {
		panic("log.Fatalf")
	})

	monkey.Patch(log.Fatal, func(...interface{}) {
		panic("log.Fatal")
	})
}

func TestGenerateDiffReport(t *testing.T) {
	t.Parallel()
	simpleMainPath := fmt.Sprintf("%s/main.go", buildExamplePackagePath(t, "simple", true))
	exampleReport := tarpReport{
		DeclaredDetails: map[string]tarpFunc{
			"A": {
				Name:     "A",
				Filename: simpleMainPath,
				DeclPos: token.Position{
					Filename: simpleMainPath,
					Offset:   16,
					Line:     3,
					Column:   1,
				},
				RBracePos: token.Position{
					Filename: simpleMainPath,
					Offset:   32,
					Line:     3,
					Column:   17,
				},
				LBracePos: token.Position{
					Filename: simpleMainPath,
					Offset:   46,
					Line:     5,
					Column:   1,
				},
			},
			"B": {
				Name:     "B",
				Filename: simpleMainPath,
				DeclPos: token.Position{
					Filename: simpleMainPath,
					Offset:   49,
					Line:     7,
					Column:   1,
				},
				RBracePos: token.Position{
					Filename: simpleMainPath,
					Offset:   65,
					Line:     7,
					Column:   17,
				},
				LBracePos: token.Position{
					Filename: simpleMainPath,
					Offset:   79,
					Line:     9,
					Column:   1,
				},
			},
			"C": {
				Name:     "C",
				Filename: simpleMainPath,
				DeclPos: token.Position{
					Filename: simpleMainPath,
					Offset:   82,
					Line:     11,
					Column:   1,
				},
				RBracePos: token.Position{
					Filename: simpleMainPath,
					Offset:   98,
					Line:     11,
					Column:   17,
				},
				LBracePos: token.Position{
					Filename: simpleMainPath,
					Offset:   112,
					Line:     13,
					Column:   1,
				},
			},
			"wrapper": {
				Name:     "wrapper",
				Filename: simpleMainPath,
				DeclPos: token.Position{
					Filename: simpleMainPath,
					Offset:   115,
					Line:     15,
					Column:   1,
				},
				RBracePos: token.Position{
					Filename: simpleMainPath,
					Offset:   130,
					Line:     15,
					Column:   16,
				},
				LBracePos: token.Position{
					Filename: simpleMainPath,
					Offset:   147,
					Line:     19,
					Column:   1,
				},
			},
		},
		Called:   set.New("A", "C", "wrapper"),
		Declared: set.New("A", "B", "C", "wrapper"),
	}

	diff := set.StringSlice(set.Difference(exampleReport.Declared, exampleReport.Called))

	expected := tarpOutput{
		LongestFunctionNameLength: 1,
		DeclaredCount:             4,
		CalledCount:               3,
		Score:                     75,
		Details: map[string][]tarpFunc{
			simpleMainPath: {
				tarpFunc{
					Name:     "B",
					Filename: simpleMainPath,
					DeclPos: token.Position{
						Filename: simpleMainPath,
						Offset:   49,
						Line:     7,
						Column:   1,
					},
					RBracePos: token.Position{
						Filename: simpleMainPath,
						Offset:   65,
						Line:     7,
						Column:   17,
					},
					LBracePos: token.Position{
						Filename: simpleMainPath,
						Offset:   79,
						Line:     9,
						Column:   1,
					},
				},
			},
		},
	}
	actual := generateDiffReport(diff, exampleReport.DeclaredDetails, exampleReport.Declared.Size(), exampleReport.Called.Size())

	assert.Equal(t, expected, actual, "expected and actual diff reports should match.")
}

func TestFuncMain(t *testing.T) {
	originalArgs := os.Args

	directoryWoes := func(t *testing.T) {
		monkey.Patch(os.Getwd, func() (string, error) {
			return "", errors.New("pineapple on pizza")
		})

		var fatalfCalled bool
		defer func() {
			if r := recover(); r != nil {
				// recovered from our monkey patched log.Fatalf
				fatalfCalled = true
			}
		}()

		os.Args = []string{
			originalArgs[0],
			"analyze",
		}

		main()
		os.Args = originalArgs
		assert.True(t, fatalfCalled, "main should call log.Fatalf() when it can't manage to retrieve the current directory")
		monkey.Unpatch(os.Getwd)
	}
	t.Run("test", directoryWoes)

	optimal := func(t *testing.T) {
		os.Args = []string{
			originalArgs[0],
			"analyze",
			fmt.Sprintf("--package=%s", buildExamplePackagePath(t, "simple", false)),
		}

		main()
		os.Args = originalArgs
	}
	t.Run("optimal", optimal)

	perfect := func(t *testing.T) {
		os.Args = []string{
			originalArgs[0],
			"analyze",
			fmt.Sprintf("--package=%s", buildExamplePackagePath(t, "perfect", false)),
		}

		main()
		os.Args = originalArgs
	}
	t.Run("perfect", perfect)

	nonexistentPackage := func(t *testing.T) {
		os.Args = []string{
			originalArgs[0],
			"analyze",
			fmt.Sprintf("--package=%s", buildExamplePackagePath(t, "absolutelynosuchpackage", false)),
			"--fail-on-found",
		}

		var fatalfCalled bool
		defer func() {
			if r := recover(); r != nil {
				// recovered from our monkey patched log.Fatalf
				fatalfCalled = true
			}
		}()

		main()
		assert.True(t, fatalfCalled, "main should call log.Fatalf() when the package dir doesn't exist")
		os.Args = originalArgs
	}
	t.Run("nonexistent package", nonexistentPackage)

	emptyPackage := func(t *testing.T) {
		os.Args = []string{
			originalArgs[0],
			"analyze",
			fmt.Sprintf("--package=%s", buildExamplePackagePath(t, "no_go_files", false)),
			"--fail-on-found",
		}

		var fatalfCalled bool
		defer func() {
			if r := recover(); r != nil {
				// recovered from our monkey patched log.Fatalf
				fatalfCalled = true
			}
		}()

		main()
		assert.True(t, fatalfCalled, "main should call log.Fatalf() when the package dir has no go files in it")
		os.Args = originalArgs
	}
	t.Run("empty package", emptyPackage)

	invalidCodeTest := func(t *testing.T) {
		os.Args = []string{
			originalArgs[0],
			"analyze",
			fmt.Sprintf("--package=%s", buildExamplePackagePath(t, "invalid", false)),
			"--fail-on-found",
		}

		invalidCodePath := buildExamplePackagePath(t, "invalid", true)
		err := os.MkdirAll(invalidCodePath, os.ModePerm)
		if err != nil {
			t.Log("error encountered creating temp path for invalid code test")
			t.FailNow()
		}

		f, err := os.Create(fmt.Sprintf("%s/main.go", invalidCodePath))
		if err != nil {
			t.Log("error encountered creating temp file for invalid code test")
			t.FailNow()
		}
		invalidCode := `
		package invalid

		import (
			"log"

		funk main() {
			return x
		)`
		fmt.Fprint(f, invalidCode)

		var fatalCalled bool
		defer func() {
			// recovered from our monkey patched log.Fatal
			if r := recover(); r != nil {
				fatalCalled = true
				err = os.RemoveAll(invalidCodePath)
				if err != nil {
					t.Logf("error encountered deleting temp directory: %v", err)
					t.FailNow()
				}
			}
		}()

		main()
		assert.True(t, fatalCalled, "main should call log.Fatal() when there is uncompilable code in the package dir")
		os.Args = originalArgs
	}
	t.Run("invalid code", invalidCodeTest)

	invalidArguments := func(t *testing.T) {
		originalArgs := os.Args

		var fatalCalled bool
		defer func() {
			// recovered from our monkey patched log.Fatal
			if r := recover(); r != nil {
				fatalCalled = true
				assert.True(t, true)
			}
		}()

		os.Args = []string{
			originalArgs[0],
			"fail plz",
		}

		main()
		os.Args = originalArgs
		assert.True(t, fatalCalled, "main should call log.Fatal when arguments are completely invalid")
	}
	t.Run("invalid arguments", invalidArguments)

	failsWhenInstructed := func(t *testing.T) {
		os.Args = []string{
			originalArgs[0],
			"analyze",
			fmt.Sprintf("--package=%s", buildExamplePackagePath(t, "simple", false)),
			"--fail-on-found",
		}
		var exitCalled bool

		monkey.Patch(os.Exit, func(code int) {
			exitCalled = true
			assert.Equal(t, 1, code, "os.Exit should be called with 1")
		})

		main()
		assert.True(t, exitCalled, "main should call log.Fatal() when --fail-on-found is passed in and extras are found")
		os.Args = originalArgs
		monkey.Unpatch(os.Exit)
	}
	t.Run("fails with --fail-on-found", failsWhenInstructed)

	padTest := func(t *testing.T) {
		failOnFound = false
		os.Args = []string{
			originalArgs[0],
			"analyze",
			fmt.Sprintf("--package=%s", buildExamplePackagePath(t, "pad_test", false)),
		}

		main()
		os.Args = originalArgs
	}
	t.Run("pad test", padTest)
}
