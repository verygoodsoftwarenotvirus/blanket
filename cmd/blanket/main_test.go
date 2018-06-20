package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/verygoodsoftwarenotvirus/blanket/analysis"
	"github.com/verygoodsoftwarenotvirus/blanket/lib/util"
	"github.com/verygoodsoftwarenotvirus/blanket/output/html"

	"github.com/bouk/monkey"
	"github.com/stretchr/testify/assert"
)

////////////////////////////////////////////////////////
//                                                    //
//                   Helper Funcs                     //
//                                                    //
////////////////////////////////////////////////////////

func buildPathForExampleFiles(t *testing.T, filename string, abs bool) string {
	t.Helper()
	gopath := os.Getenv("GOPATH")
	if abs {
		return fmt.Sprintf("%s/src/github.com/verygoodsoftwarenotvirus/blanket/example_files/%s", gopath, filename)
	}
	return fmt.Sprintf("github.com/verygoodsoftwarenotvirus/blanket/example_files/%s", filename)
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

func TestFuncMain(t *testing.T) {
	originalArgs := os.Args

	t.Run("test", func(_t *testing.T) {
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
	})

	t.Run("optimal", func(_t *testing.T) {
		os.Args = []string{
			originalArgs[0],
			"analyze",
			fmt.Sprintf("--package=%s", util.BuildExamplePackagePath(t, "simple", false)),
		}

		main()
		os.Args = originalArgs
	})

	t.Run("perfect", func(_t *testing.T) {
		os.Args = []string{
			originalArgs[0],
			"analyze",
			fmt.Sprintf("--package=%s", util.BuildExamplePackagePath(t, "perfect", false)),
		}

		main()
		os.Args = originalArgs
	})

	t.Run("package as argument", func(_t *testing.T) {
		os.Args = []string{
			originalArgs[0],
			"analyze",
			util.BuildExamplePackagePath(t, "perfect", false),
		}

		main()
		os.Args = originalArgs
	})

	t.Run("nonexistent package", func(_t *testing.T) {
		os.Args = []string{
			originalArgs[0],
			"analyze",
			fmt.Sprintf("--package=%s", util.BuildExamplePackagePath(t, "absolutelynosuchpackage", false)),
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
	})

	t.Run("empty package", func(_t *testing.T) {
		os.Args = []string{
			originalArgs[0],
			"analyze",
			fmt.Sprintf("--package=%s", util.BuildExamplePackagePath(t, "no_go_files", false)),
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
	})

	t.Run("invalid code", func(_t *testing.T) {
		os.Args = []string{
			originalArgs[0],
			"analyze",
			fmt.Sprintf("--package=%s", util.BuildExamplePackagePath(t, "invalid", false)),
			"--fail-on-found",
		}

		invalidCodePath := util.BuildExamplePackagePath(t, "invalid", true)
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
	})

	t.Run("invalid arguments", func(_t *testing.T) {
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
	})

	t.Run("fails with --fail-on-found", func(_t *testing.T) {
		os.Args = []string{
			originalArgs[0],
			"analyze",
			fmt.Sprintf("--package=%s", util.BuildExamplePackagePath(t, "simple", false)),
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
	})

	t.Run("pad test", func(_t *testing.T) {
		failOnFound = false
		os.Args = []string{
			originalArgs[0],
			"analyze",
			fmt.Sprintf("--package=%s", util.BuildExamplePackagePath(t, "pad_test", false)),
		}

		main()
		os.Args = originalArgs
	})

	t.Run("JSON test", func(_t *testing.T) {
		failOnFound = false
		os.Args = []string{
			originalArgs[0],
			"analyze",
			"--json",
			fmt.Sprintf("--package=%s", util.BuildExamplePackagePath(t, "pad_test", false)),
		}

		main()
		os.Args = originalArgs
	})

	t.Run("basic cover test", func(_t *testing.T) {
		monkey.Patch(html.StartBrowser, func(url, os string) bool { return true })
		os.Args = []string{
			originalArgs[0],
			"cover",
			fmt.Sprintf(
				"--html=%s",
				buildPathForExampleFiles(_t, "simple_count.coverprofile", true),
			),
		}

		main()
		os.Args = originalArgs
		monkey.Unpatch(html.StartBrowser)
	})

	t.Run("cover fails when it cannot parse the profile", func(_t *testing.T) {
		var fatalCalled bool
		defer func() {
			// recovered from our monkey patched log.Fatal
			r := recover()
			if x, ok := r.(string); ok {
				if x == "log.Fatal" {
					fatalCalled = true
				}
			}
		}()

		os.Args = []string{
			originalArgs[0],
			"cover",
			`--html=""`,
		}

		main()
		assert.True(t, fatalCalled)
		os.Args = originalArgs
	})

	t.Run("cover fails when it cannot generate HTML output", func(_t *testing.T) {
		monkey.Patch(html.Output, func(string, string, *analysis.BlanketReport) error { return errors.New("pineapple on pizza") })

		var fatalCalled bool
		defer func() {
			// recovered from our monkey patched log.Fatal
			if r := recover(); r != nil {
				fatalCalled = true
			}
		}()

		os.Args = []string{
			originalArgs[0],
			"cover",
			fmt.Sprintf(
				"--html=%s",
				buildPathForExampleFiles(_t, "simple_count.coverprofile", true),
			),
		}

		main()
		assert.True(t, fatalCalled)
		os.Args = originalArgs
		monkey.Unpatch(html.Output)
	})
}
