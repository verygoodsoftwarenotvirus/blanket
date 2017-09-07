package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"

	"github.com/fatih/set"
	"github.com/jessevdk/go-flags"
)

func getDeclaredNames(in *ast.File, out *set.Set) {
	for _, x := range in.Decls {
		switch f := x.(type) {
		case *ast.FuncDecl:
			functionName := f.Name.Name // "Avoid Stutter" lol
			var parentName string
			if f.Recv != nil && len(f.Recv.List) == 1 {
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

func main() {
	gopath := os.Getenv("GOPATH")
	var opts struct {
		Package      string `short:"p" long:"package" required:"true" description:"the package to analyze"`
		FailOnExtras bool   `short:"f" long:"fail-on-extras" description:"exit 1 when uncalled functions are found"`
	}
	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}

	pkgDir := strings.Join([]string{gopath, "src", opts.Package}, "/")

	_, err = os.Stat(pkgDir)
	if os.IsNotExist(err) {
		log.Fatalf("packageDir doesn't exist: %s", pkgDir)
	}

	astPkg, err := parser.ParseDir(token.NewFileSet(), pkgDir, nil, parser.AllErrors)
	if err != nil {
		log.Fatal(err)
	}

	declaredFuncs := set.New()
	calledFuncs := set.New()

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

	if opts.FailOnExtras && len(diff) > 0 {
		errorString := fmt.Sprintf(`The following functions are declared but not called in any tests:
	%s
		`, strings.Join(diff, ",\n\t"))
		log.Fatal(errorString)
	}
}
