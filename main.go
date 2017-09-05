package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"

	"github.com/fatih/set"
	"strings"
)

func use(...interface{}) {
	return
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
	pkg, err := parser.ParseDir(token.NewFileSet(), "example_files/simple", nil, parser.AllErrors)
	if err != nil {
		log.Fatal(err)
	}

	declaredFuncs := set.New()
	calledFuncs := set.New()

	for name, f := range pkg["simple"].Files {
		isTest := strings.HasSuffix(name, "_test.go")
		if isTest {
			getCalledNames(f, calledFuncs)
		} else {
			getDeclaredNames(f, declaredFuncs)
		}
	}

	difference := set.Difference(declaredFuncs, calledFuncs)
	diff := set.StringSlice(difference)

	log.Println(diff)
}
