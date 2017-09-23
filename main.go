// +build !unit

package main

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"log"
	"os"
	"sort"
	"strings"
	"text/template"
	"unicode/utf8"

	"github.com/fatih/set"
	"github.com/spf13/cobra"
)

const (
	differenceReportTmpl = `The following functions are declared, but don't appear to have direct unit tests:{{range $filename, $missing := .}}
in {{$filename}}:{{range $missing}}
	{{pad .Name}} on line {{.DeclPos.Line}}{{end}}{{end}}
`
)

var (
	// flags
	debug          bool
	failOnFinding  bool
	analyzePackage string

	// helper variables
	fileset *token.FileSet
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
	Run: func(cmd *cobra.Command, args []string) {
		analyze(analyzePackage, failOnFinding)
	},
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
	rootCmd.Flags().BoolVarP(&debug, "debug", "d", false, "log various details about the parsing process")
	analyzeCmd.Flags().BoolVarP(&failOnFinding, "fail-on-found", "f", false, "Call os.Exit(1) when functions without direct tests are found")
	analyzeCmd.Flags().StringVarP(&analyzePackage, "package", "p", ".", "Package to run analyze on. Defaults to the current directory.")

	fileset = token.NewFileSet()
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

	astPkg, err := parser.ParseDir(fileset, pkgDir, nil, parser.AllErrors)
	if err != nil {
		log.Fatal(err)
	}

	if len(astPkg) == 0 || astPkg == nil {
		log.Fatal("no go files found!")
	}

	declaredFuncInfo := map[string]TarpFunc{}
	calledFuncs := set.New("init")

	for _, pkg := range astPkg {
		for name, f := range pkg.Files {
			if debug {
				log.Printf("parsing %s", name)
			}
			isTest := strings.HasSuffix(name, "_test.go")
			if isTest {
				getCalledNames(f, calledFuncs)
			} else {
				getDeclaredNames(f, fileset, declaredFuncInfo)
			}
		}
	}

	declaredFuncs := set.New()
	for _, f := range declaredFuncInfo {
		declaredFuncs.Add(f.Name)
	}

	diff := set.StringSlice(set.Difference(declaredFuncs, calledFuncs))
	diffReport := generateDiffReport(diff, declaredFuncInfo)

	if len(diff) > 0 {
		fmt.Println(diffReport)
		if failOnFinding {
			os.Exit(1)
		}
	}
}

func generateDiffReport(diff []string, declaredFuncInfo map[string]TarpFunc) string {
	longestFunctionNameLength := 0
	missingFuncs := &TarpDetails{}
	for _, s := range diff {
		if utf8.RuneCountInString(s) > longestFunctionNameLength {
			longestFunctionNameLength = len(s)
		}
		*missingFuncs = append(*missingFuncs, declaredFuncInfo[s])
	}
	sort.Sort(missingFuncs)
	byFilename := map[string][]TarpFunc{}
	for _, tf := range *missingFuncs {
		byFilename[tf.Filename] = append(byFilename[tf.Filename], tf)
	}

	funcMap := template.FuncMap{
		// The name "title" is what the function will be called in the template text.
		"pad": func(s string) string {
			// https://github.com/willf/pad/blob/master/pad.go
			numberOfSpacesToAdd := longestFunctionNameLength - utf8.RuneCountInString(s)
			for i := 0; i < numberOfSpacesToAdd; i++ {
				s += " "
			}
			return s
		},
	}

	t, err := template.New("t").Funcs(funcMap).Parse(differenceReportTmpl)
	if err != nil {
		panic(err)
	}

	var tpl bytes.Buffer
	if err = t.Execute(&tpl, byFilename); err != nil {
		panic(err)
	}
	return tpl.String()
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
