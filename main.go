// +build !unit

package main

import (
	"bytes"
	"fmt"
	"go/token"
	"log"
	"sort"
	"strconv"
	"text/template"
	"unicode/utf8"

	"github.com/fatih/color"
	"github.com/fatih/set"
	"github.com/spf13/cobra"
	"os"
)

const (
	differenceReportTmpl = `{{$len := .LongestFunctionNameLength}}Functions without direct unit tests:{{range $filename, $missing := .Details}}
in {{colorizer $filename "white" true}}:{{range $missing}}
	{{pad .Name $len}} on line {{.DeclPos.Line}}{{end}}{{end}}

Grade: {{grader .Score}} ({{.CalledCount}}/{{.DeclaredCount}} functions)
`
	perfectScoreTmpl = `Grade: {{grader .Score}} ({{.CalledCount}}/{{.DeclaredCount}} functions)`
)

var (
	// flags
	failOnFound    bool
	debug 			bool
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
			report := analyze(analyzePackage)
			diff := set.StringSlice(set.Difference(report.Declared, report.Called))
			diffReport := generateDiffReport(diff, report.DeclaredDetails, report.Declared.Size(), report.Called.Size())

			var templateToUse string
			if len(diff) > 0 {
				templateToUse = differenceReportTmpl
			} else {
				templateToUse = perfectScoreTmpl
			}

			var tpl bytes.Buffer
			// see above re: the error this function returns
			t, _ := template.New("t").Funcs(templateFuncMap).Parse(templateToUse)
			t.Execute(&tpl, diffReport)
			fmt.Println(tpl.String())

			if len(diff) > 0 && failOnFound {
				os.Exit(1)
			}
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

	templateFuncMap = template.FuncMap{
		"pad": func(s string, longest int) string {
			// https://github.com/willf/pad/blob/master/pad.go
			numberOfSpacesToAdd := longest - utf8.RuneCountInString(s)
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
)

func init() {
	rootCmd.AddCommand(analyzeCmd)
	analyzeCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Call os.Exit(1) when functions without direct tests are found")
	analyzeCmd.Flags().BoolVarP(&failOnFound, "fail-on-found", "f", false, "Call os.Exit(1) when functions without direct tests are found")
	analyzeCmd.Flags().StringVarP(&analyzePackage, "package", "p", ".", "Package to run analyze on. Defaults to the current directory.")

	fileset = token.NewFileSet()
}

func generateDiffReport(diff []string, declaredFuncInfo map[string]TarpFunc, declaredFuncCount int, calledFuncCount int) TarpOutput {
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

	report := TarpOutput{
		DeclaredCount: declaredFuncCount,
		CalledCount:   calledFuncCount,
		Score:         int(score * 100),
		Details:       byFilename,
		LongestFunctionNameLength: longestFunctionNameLength,
	}

	return report
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
