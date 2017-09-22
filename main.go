// +build !unit

package main

import (
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"

	"github.com/fatih/set"
	"github.com/spf13/cobra"
)

var (
	failOnFinding  bool
	analyzePackage string
)

var rootCmd = &cobra.Command{
	Use:   "tarp",
	Short: "tarp is a coverage helper tool",
	Long:  `tarp is a tool which aims to help ensure you have direct unit tests for all your declared functions for a particular Go package.`,
}

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze a given package",
	Long:  "Analyze takes a given package and determines which functions lack direct unit tests.",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if analyzePackage == "" {
			return errors.New("the required flag `-p, --package` was not specified")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		analyze(analyzePackage, failOnFinding)
	},
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
	analyzeCmd.Flags().BoolVarP(&failOnFinding, "fail-on-found", "f", false, "Call os.Exit(1) when functions without direct tests are found")
	analyzeCmd.Flags().StringVarP(&analyzePackage, "package", "p", ".", "Package to run analyze on. Defaults to the current directory.")
}

func analyze(analyzePackage string, failOnFinding bool) {
	gopath := os.Getenv("GOPATH")

	pkgDir := strings.Join([]string{gopath, "src", analyzePackage}, "/")
	if analyzePackage == "." {
		var err error
		pkgDir, err = os.Getwd()
		if err != nil {
			log.Fatalf("error encountered getting current working directory: %v", err)
		}
	}

	_, err := os.Stat(pkgDir)
	if os.IsNotExist(err) {
		log.Fatalf("packageDir doesn't exist: %s", pkgDir)
	}

	astPkg, err := parser.ParseDir(token.NewFileSet(), pkgDir, nil, parser.AllErrors)
	if err != nil {
		log.Fatal(err)
	}

	declaredFuncs := set.New()
	calledFuncs := set.New("init")

	if len(astPkg) == 0 || astPkg == nil {
		log.Fatal("no go files found!")
	}

	for _, pkg := range astPkg {
		for name, f := range pkg.Files {
			isTest := strings.HasSuffix(name, "_test.go")
			if isTest {
				getCalledNames(f, calledFuncs)
			} else {
				getDeclaredNames(f, declaredFuncs)
			}
		}
	}
	diff := set.StringSlice(set.Difference(declaredFuncs, calledFuncs))
	diffReport := fmt.Sprintf(`The following functions are declared but not called in any tests:
	%s
	`, strings.Join(diff, ",\n\t"))

	if len(diff) > 0 {
		if failOnFinding {
			log.Fatal(diffReport)
		} else {
			log.Println(diffReport)
		}
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
