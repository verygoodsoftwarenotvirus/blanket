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

func use(...interface{}) {
	return
}

func dirExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

func getDeclaredNames(in *ast.File, out *set.Set) {
	for _, x := range in.Decls {
		switch n := x.(type) {
		case *ast.FuncDecl:
			out.Add(n.Name.Name) // "Avoid Stutter" lol
		}
	}
}

func getCalledNames(in *ast.File, out *set.Set) {
	// Using switches here to avoid panics, this is probably wrong and bad but ¯\_(ツ)_/¯
	for _, x := range in.Decls {
		switch n := x.(type) {
		case *ast.FuncDecl:
			for _, le := range n.Body.List {
				switch e := le.(type) {
				case *ast.ExprStmt:
					switch c := e.X.(type) {
					case *ast.CallExpr:
						switch f := c.Fun.(type) {
						case *ast.Ident:
							out.Add(f.Name)
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

	difference := set.Difference(declaredFuncs, calledFuncs)
	diff := set.StringSlice(difference)

	if opts.FailOnExtras && len(diff) > 0 {
		errorString := fmt.Sprintf(`The following functions are declared but not called in any tests:
	%s
		`, strings.Join(diff, ",\n\t"))
		log.Fatal(errorString)
	}
}
