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
	"strconv"
	"strings"
	"text/template"
	"unicode/utf8"

	"github.com/fatih/color"
	"github.com/fatih/set"
	"github.com/spf13/cobra"
)

const (
	differenceReportTmpl = `Functions without direct unit tests:{{range $filename, $missing := .Details}}
in {{colorizer $filename "white" true}}:{{range $missing}}
	{{pad .Name}} on line {{.DeclPos.Line}}{{end}}{{end}}

Grade: {{grader .Score}} ({{.CalledCount}}/{{.DeclaredCount}} functions)
`
)

var (
	// flags
	failOnFound    bool
	analyzePackage string

	// helper variables
	fileset *token.FileSet

	// commands
	rootCmd = &cobra.Command{
		Use:   "tarp",
		Short: "tarp is a coverage helper tool",
		Long:  `tarp is a tool which aims to help ensure you have direct unit tests for all your declared functions for a particular Go package.`,
	}

	analyzeCmd = &cobra.Command{
		Use:   "analyze",
		Short: "Analyze a given package",
		Long:  "Analyze takes a given package and determines which functions lack direct unit tests.",
		Run: func(cmd *cobra.Command, args []string) {
			analyze(analyzePackage, failOnFound)
		},
	}

	colors = map[string]color.Attribute{
		"black":   color.FgBlack,
		"red":     color.FgRed,
		"green":   color.FgGreen,
		"yellow":  color.FgYellow,
		"blue":    color.FgBlue,
		"magenta": color.FgMagenta,
		"cyan":    color.FgCyan,
		"white":   color.FgWhite,
	}
)

func init() {
	rootCmd.AddCommand(analyzeCmd)
	analyzeCmd.Flags().BoolVarP(&failOnFound, "fail-on-found", "f", false, "Call os.Exit(1) when functions without direct tests are found")
	analyzeCmd.Flags().StringVarP(&analyzePackage, "package", "p", ".", "Package to run analyze on. Defaults to the current directory.")

	fileset = token.NewFileSet()
}

func generateDiffReport(diff []string, declaredFuncInfo map[string]TarpFunc, declaredFuncCount int, calledFuncCount int) string {
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
	score := float64(calledFuncCount) / float64(declaredFuncCount)

	report := struct {
		DeclaredCount int
		CalledCount   int
		Score         int
		Details       map[string][]TarpFunc
	}{
		DeclaredCount: declaredFuncCount,
		CalledCount:   calledFuncCount,
		Score:         int(score * 100),
		Details:       byFilename,
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
		"colorizer": func(s string, c string, bold bool) string {
			arguments := []color.Attribute{colors[c]}
			if bold {
				arguments = append(arguments, color.Bold)
			}
			return color.New(arguments...).SprintfFunc()(s)
		},
		"grader": func(score int) string {
			gradeMap := map[int]string{
				6:  "magenta",
				7:  "yellow",
				8:  "cyan",
				9:  "blue",
				10: "green",
			}

			grade := "red"
			if realGrade, ok := gradeMap[score/10]; ok {
				grade = realGrade
			}
			return color.New(colors[grade]).SprintfFunc()(strconv.Itoa(score) + "%%")
		},
	}

	// ignoring the error here because we can't recreate it in tests because our values are already fine.
	t, _ := template.New("t").Funcs(funcMap).Parse(differenceReportTmpl)

	var tpl bytes.Buffer
	// see above re: the error this function returns
	t.Execute(&tpl, report)
	return tpl.String()
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
	helperFunctionReturnMap := map[string][]string{}
	nameToTypeMap := map[string]string{}

	for _, pkg := range astPkg {
		for name, f := range pkg.Files {
			if strings.HasSuffix(name, "_test.go") {
				getCalledNames(f, nameToTypeMap, helperFunctionReturnMap, calledFuncs)
			} else {
				getDeclaredNames(f, fileset, declaredFuncInfo)
			}
		}
	}

	declaredFuncs := set.New()
	for _, f := range declaredFuncInfo {
		declaredFuncs.Add(f.Name)
	}
	toPrune := set.StringSlice(set.Difference(calledFuncs, declaredFuncs))
	for _, x := range toPrune {
		calledFuncs.Remove(x)
	}

	diff := set.StringSlice(set.Difference(declaredFuncs, calledFuncs))
	diffReport := generateDiffReport(diff, declaredFuncInfo, declaredFuncs.Size(), calledFuncs.Size())

	if len(diff) > 0 {
		fmt.Println(diffReport)
		if failOnFinding {
			os.Exit(1)
		}
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
