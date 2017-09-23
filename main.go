// +build !unit

package main

import (
	"go/parser"
	"go/token"
	"sort"
	"text/template"
	"bytes"
	"unicode/utf8"
	"log"
	"os"
	"strings"

	"github.com/fatih/set"
	"github.com/spf13/cobra"
	"fmt"
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

	fileset := token.NewFileSet()
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
	for _, tf := range *missingFuncs{
		if _, ok := byFilename[tf.Filename]; !ok {
			byFilename[tf.Filename] = []TarpFunc{tf}
		} else {
			byFilename[tf.Filename] = append(byFilename[tf.Filename], tf)
		}
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

	differenceReportTmpl := `The following functions are declared, but don't appear to have direct unit tests:{{range $filename, $missing := .}}
{{$filename}}:{{range $missing}}
	{{pad .Name}} on line {{.DeclPos.Line}}{{end}}{{end}}
`
	t := template.New("t").Funcs(funcMap)
	t, err := t.Parse(differenceReportTmpl)
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
