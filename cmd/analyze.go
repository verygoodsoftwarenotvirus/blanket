package tarp

import (
	"errors"
	"fmt"
	"go/ast"
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

func init() {
	RootCmd.AddCommand(analyzeCmd)
	analyzeCmd.Flags().BoolVarP(&failOnFinding, "fail-on-finding", "f", false, "Call os.Exit(1) when functions without direct tests are found")
	analyzeCmd.Flags().StringVarP(&analyzePackage, "package", "p", "", "Package to run analyze on")
}

func getDeclaredNames(in *ast.File, out *set.Set) {
	for _, x := range in.Decls {
		switch f := x.(type) {
		case *ast.FuncDecl:
			functionName := f.Name.Name // "Avoid Stutter" lol
			var parentName string
			if f.Recv != nil && len(f.Recv.List) == 1 { // handles things like `type Example struct`
				r := f.Recv.List[0]
				parentName = r.Type.(*ast.StarExpr).X.(*ast.Ident).Name
			}

			if parentName != "" {
				out.Add(fmt.Sprintf("%s.%s", parentName, functionName))
			} else {
				out.Add(functionName)
			}
		}
	}
}

func getCalledNames(in *ast.File, out *set.Set) {
	// Using switches here to avoid panics, this is probably wrong and bad but ¯\_(ツ)_/¯
	nameToTypeMap := map[string]string{}
	for _, x := range in.Decls {
		switch n := x.(type) {
		case *ast.GenDecl:
			for _, spec := range n.Specs {
				switch global := spec.(type) {
				case *ast.ValueSpec: // for things like `var e Example` declared outside of functions
					varName := global.Names[0].Name
					typeName := global.Type.(*ast.Ident).Name
					nameToTypeMap[varName] = typeName
				}
			}
		case *ast.FuncDecl:
			for _, le := range n.Body.List {
				switch e := le.(type) {
				case *ast.AssignStmt: // handles things like `e := Example{}` (with or without &)
					varName := e.Lhs[0].(*ast.Ident).Name
					typeName := e.Rhs[0].(*ast.UnaryExpr).X.(*ast.CompositeLit).Type.(*ast.Ident).Name
					nameToTypeMap[varName] = typeName
				case *ast.DeclStmt: // handles things like `var e Example`
					varName := e.Decl.(*ast.GenDecl).Specs[0].(*ast.ValueSpec).Names[0].Name
					typeName := e.Decl.(*ast.GenDecl).Specs[0].(*ast.ValueSpec).Type.(*ast.Ident).Name
					nameToTypeMap[varName] = typeName
				case *ast.ExprStmt: // handles function calls
					switch c := e.X.(type) {
					case *ast.CallExpr:
						switch f := c.Fun.(type) {
						case *ast.Ident:
							out.Add(f.Name)
						case *ast.SelectorExpr:
							structVarName := f.X.(*ast.Ident).Name
							calledMethodName := f.Sel.Name
							out.Add(fmt.Sprintf("%s.%s", nameToTypeMap[structVarName], calledMethodName))
						}
					}
				}
			}
		}
	}
}

func buildPackagePath(path string) string {
	gopath := os.Getenv("GOPATH")
	pkgDir := strings.Join([]string{gopath, "src", path}, "/")
	return pkgDir
}

func analyze(_ *cobra.Command, _ []string) {
	pkgDir := buildPackagePath(analyzePackage)
	log.Println("analyzing ", pkgDir)

	_, err := os.Stat(pkgDir)
	if os.IsNotExist(err) {
		log.Fatalf("package dir doesn't exist: %s", pkgDir)
	}

	astPkg, err := parser.ParseDir(token.NewFileSet(), pkgDir, nil, parser.AllErrors)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(len(astPkg))
	declaredFuncs := set.New()
	calledFuncs := set.New()

	for _, pkg := range astPkg {
		for name, f := range pkg.Files {
			log.Println("analyzing ", name)
			isTest := strings.HasSuffix(name, "_test.go")
			if isTest {
				getCalledNames(f, calledFuncs)
			} else {
				getDeclaredNames(f, declaredFuncs)
			}
		}
	}

	if true {
		declaredFuncNames := set.StringSlice(declaredFuncs)
		log.Printf(`declared functions:
			%s
		`, strings.Join(declaredFuncNames, "\n\t"))

		calledFuncNames := set.StringSlice(calledFuncs)
		log.Printf(`called functions:
			%s
		`, strings.Join(calledFuncNames, "\n\t"))
	}

	diff := set.StringSlice(set.Difference(declaredFuncs, calledFuncs))

	if failOnFinding && len(diff) > 0 {
		errorString := fmt.Sprintf(`The following functions are declared but not called in any tests:
%s
	`, strings.Join(diff, ",\n\t"))
		log.Fatal(errorString)
	}
}

// analyzeCmd represents the analyze command
var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze a given package",
	Long:  "Analyze takes a given package's code and determines which functions lack direct test coverage.",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if analyzePackage == "" {
			return errors.New("the required flag `-p, --package` was not specified")
		}
		return nil
	},
	Run: analyze,
}
