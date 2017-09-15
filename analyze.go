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
)

func parseCallExpr(in *ast.CallExpr, nameToTypeMap map[string]string, out *set.Set) {
	switch f := in.Fun.(type) {
	case *ast.Ident:
		out.Add(f.Name)
	case *ast.SelectorExpr:
		structVarName := f.X.(*ast.Ident).Name
		calledMethodName := f.Sel.Name
		out.Add(fmt.Sprintf("%s.%s", nameToTypeMap[structVarName], calledMethodName))
	}
}

func parseUnaryExpr(in *ast.UnaryExpr, varName string, nameToTypeMap map[string]string) {
	switch u := in.X.(*ast.CompositeLit).Type.(type) {
	case *ast.Ident:
		nameToTypeMap[varName] = u.Name
	}
}

func parseDeclStmt(in *ast.DeclStmt, nameToTypeMap map[string]string) {
	varName := in.Decl.(*ast.GenDecl).Specs[0].(*ast.ValueSpec).Names[0].Name
	typeName := in.Decl.(*ast.GenDecl).Specs[0].(*ast.ValueSpec).Type.(*ast.Ident).Name
	nameToTypeMap[varName] = typeName
}

func parseExprStmt(in *ast.ExprStmt, nameToTypeMap map[string]string, out *set.Set) {
	switch c := in.X.(type) {
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

func parseGenDecl(in *ast.GenDecl, nameToTypeMap map[string]string) {
	for _, spec := range in.Specs {
		switch global := spec.(type) {
		case *ast.ValueSpec: // for things like `var e Example` declared outside of functions
			if len(global.Names) > 1 {
				log.Println("wtf")
			}

			varName := global.Names[0].Name
			if global.Type != nil {
				typeName := global.Type.(*ast.Ident).Name
				nameToTypeMap[varName] = typeName
			}
		}
	}
}

func parseFuncDecl(f *ast.FuncDecl, out *set.Set) {
	functionName := f.Name.Name // "Avoid Stutter" lol
	var parentName string
	if f.Recv != nil {
		parentName = f.Recv.List[0].Type.(*ast.StarExpr).X.(*ast.Ident).Name
	}

	if parentName != "" {
		out.Add(fmt.Sprintf("%s.%s", parentName, functionName))
	} else {
		out.Add(functionName)
	}
}

func getDeclaredNamesFromFile(in *ast.File, out *set.Set) {
	for _, x := range in.Decls {
		getDeclaredNames(x, out)
	}
}

func getDeclaredNames(in ast.Decl, out *set.Set) {
	switch f := in.(type) {
	case *ast.FuncDecl:
		parseFuncDecl(f, out)
	}
}

func getCalledNamesFromFunctionLiteral(in *ast.FuncLit, nameToTypeMap map[string]string, out *set.Set) {
	for _, le := range in.Body.List {
		switch e := le.(type) {
		case *ast.AssignStmt: // handles things like `e := Example{}` (with or without &)
			varName := e.Lhs[0].(*ast.Ident).Name
			switch t := e.Rhs[0].(type) {
			case *ast.UnaryExpr:
				parseUnaryExpr(t, varName, nameToTypeMap)
			case *ast.CallExpr:
				parseCallExpr(t, nameToTypeMap, out)
			}

		case *ast.DeclStmt: // handles things like `var e Example`
			parseDeclStmt(e, nameToTypeMap)
		case *ast.ExprStmt: // handles function calls
			parseExprStmt(e, nameToTypeMap, out)
		}
	}
}

func parseStmt(in ast.Stmt, nameToTypeMap map[string]string, out *set.Set) {
		switch e := in.(type) {
		case *ast.AssignStmt: // handles things like `e := Example{}` (with or without &)
			varName := e.Lhs[0].(*ast.Ident).Name
			switch t := e.Rhs[0].(type) {
			case *ast.FuncLit:
				getCalledNamesFromFunctionLiteral(t, nameToTypeMap, out)
			case *ast.UnaryExpr:
				parseUnaryExpr(t, varName, nameToTypeMap)
			case *ast.CallExpr:
				parseCallExpr(t, nameToTypeMap, out)
			}
		case *ast.RangeStmt:
			for _, x := range e.Body.List {
				parseStmt(x, nameToTypeMap, out)
			}
		case *ast.IfStmt:
			for _, x := range e.Body.List{
				parseStmt(x, nameToTypeMap, out)
			}
		case *ast.DeclStmt: // handles things like `var e Example`
			parseDeclStmt(e, nameToTypeMap)
		case *ast.ExprStmt: // handles function calls
			parseExprStmt(e, nameToTypeMap, out)
		}
}

func getCalledNames(in *ast.File, out *set.Set) {
	// Using switches here to avoid panics, this is probably wrong and bad but ¯\_(ツ)_/¯
	nameToTypeMap := map[string]string{}
	for _, d := range in.Decls {
		switch n := d.(type) {
		case *ast.GenDecl:
			parseGenDecl(n, nameToTypeMap)
		case *ast.FuncDecl:
			// functionName := n.Name.Name
			for _, le := range n.Body.List {
				parseStmt(le, nameToTypeMap, out)
			}
		}
	}
}

func analyze(analyzePackage string, failOnFinding bool) {
	gopath := os.Getenv("GOPATH")
	pkgDir := strings.Join([]string{gopath, "src", analyzePackage}, "/")

	_, err := os.Stat(pkgDir)
	if os.IsNotExist(err) {
		log.Fatalf("packageDir doesn't exist: %s", pkgDir)
	}

	astPkg, err := parser.ParseDir(token.NewFileSet(), pkgDir, nil, parser.AllErrors)
	if err != nil {
		log.Fatal(err)
	}

	declaredFuncs := set.New()
	calledFuncs := set.New()

	if len(astPkg) == 0 {
		log.Fatal("no go files found!")
	}

	for _, pkg := range astPkg {
		for name, f := range pkg.Files {
			isTest := strings.HasSuffix(name, "_test.go")
			if isTest {
				getCalledNames(f, calledFuncs)
			} else {
				getDeclaredNamesFromFile(f, declaredFuncs)
			}
		}
	}
	diff := set.StringSlice(set.Difference(declaredFuncs, calledFuncs))
	diffReport := fmt.Sprintf(`The following functions are declared but not called in any tests:
%s
	`, strings.Join(diff, ",\n\t"))

	if len(diff) > 0 {
		if failOnFinding{
			log.Fatal(diffReport)
		} else {
			log.Println(diffReport)
		}
	}
}
