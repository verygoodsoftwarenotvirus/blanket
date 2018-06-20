// blanket finds functions without direct unit tests
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"unicode/utf8"

	"github.com/verygoodsoftwarenotvirus/blanket/analysis"
	"github.com/verygoodsoftwarenotvirus/blanket/output/html"

	"github.com/fatih/color"
	"github.com/fatih/set"
	"github.com/spf13/cobra"
	"golang.org/x/tools/cover"
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
	// global flags
	debug   bool
	verbose bool

	// analyze flags
	failOnFound    bool
	outputAsJSON   bool
	analyzePackage string

	// cover flags
	coverprofile string

	// helper variables
	fileset *token.FileSet

	// commands
	rootCmd = &cobra.Command{
		Use:   "blanket",
		Short: "blanket is a coverage helper tool",
		Long:  `blanket is a tool that helps you catch functions which don't have direct unit tests in your Go libraries`,
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
			return fmt.Sprintf("%s%s", strings.Repeat(" ", longest-utf8.RuneCountInString(s)), s)
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

	analyzeCmd = &cobra.Command{
		Use:   "analyze",
		Short: "Analyze a given package",
		Long:  "Analyze takes a given package and determines which functions lack direct unit tests.",
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: figure out why tests won't capture coverage for this
			// if len(args) == 1 && analyzePackage == "." {
			// 		analyzePackage = args[0]
			// }
			analyzer := analysis.NewAnalyzer()
			report, err := analyzer.Analyze(analyzePackage)
			if err != nil {
				log.Fatal(err)
			}

			diff := set.StringSlice(set.Difference(report.Declared, report.Called))
			diffReport := analyzer.GenerateDiffReport()

			if outputAsJSON {
				json.NewEncoder(os.Stdout).Encode(diffReport)
			} else {
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
			}

			if len(diff) > 0 && failOnFound {
				os.Exit(1)
			}
		},
	}

	coverCmd = &cobra.Command{
		Use:   "cover",
		Short: "Open a web browser displaying annotated source code",
		Long:  "Cover takes a given coverprofile and produces HTML with coverage info",
		Run: func(cmd *cobra.Command, args []string) {
			profiles, err := cover.ParseProfiles(coverprofile)
			if err != nil {
				log.Fatal(err)
			}
			pkgPath := filepath.Dir(profiles[0].FileName)

			analyzer := analysis.NewAnalyzer()
			report, err := analyzer.Analyze(pkgPath)
			if err != nil {
				log.Fatal(err)
			}

			err = html.Output(coverprofile, "", report)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "log select debug information")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	fileset = token.NewFileSet()

	rootCmd.AddCommand(analyzeCmd)
	analyzeCmd.Flags().BoolVarP(&outputAsJSON, "json", "j", false, "Render results as a JSON blob")
	analyzeCmd.Flags().BoolVarP(&failOnFound, "fail-on-found", "F", false, "Call os.Exit(1) when functions without direct tests are found")
	analyzeCmd.Flags().StringVarP(&analyzePackage, "package", "p", ".", "Package to run analyze on. Defaults to the current directory.")

	rootCmd.AddCommand(coverCmd)
	coverCmd.Flags().StringVarP(&coverprofile, "html", "c", "", "coverprofile to generate HTML for.")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
