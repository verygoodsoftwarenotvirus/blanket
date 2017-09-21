package main

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/fatih/set"
)

func parseCallExpr(in *ast.CallExpr, nameToTypeMap map[string]string, helperFunctionReturnMap map[string][]string, out *set.Set) {
	// FIXME: iterate over in.Args to see if there are function calls
	switch f := in.Fun.(type) {
	case *ast.Ident:
		functionName := f.Name
		if _, ok := helperFunctionReturnMap[functionName]; !ok {
			out.Add(functionName)
		}
	case *ast.SelectorExpr:
		structVarName := f.X.(*ast.Ident).Name
		calledMethodName := f.Sel.Name
		if _, ok := nameToTypeMap[structVarName]; ok {
			out.Add(fmt.Sprintf("%s.%s", nameToTypeMap[structVarName], calledMethodName))
		}
	}
}

// parseUnaryExpr parses Unary expressions. From the go/ast docs:
//      A UnaryExpr node represents a unary expression. Unary "*" expressions are represented via StarExpr nodes.
// (handles declarations like `callExpr := &ast.CallExpr`)
func parseUnaryExpr(in *ast.UnaryExpr, varName string, nameToTypeMap map[string]string) {
	switch u := in.X.(*ast.CompositeLit).Type.(type) {
	case *ast.Ident:
		nameToTypeMap[varName] = u.Name
	}
}

// parseDeclStmt parses declaration statments. From the go/ast docs:
// 		A DeclStmt node represents a declaration in a statement list.
// DeclStmts come from function bodies, GenDecls come from package-wide const or var declarations
func parseDeclStmt(in *ast.DeclStmt, nameToTypeMap map[string]string) {
	varName := in.Decl.(*ast.GenDecl).Specs[0].(*ast.ValueSpec).Names[0].Name
	typeName := in.Decl.(*ast.GenDecl).Specs[0].(*ast.ValueSpec).Type.(*ast.Ident).Name
	nameToTypeMap[varName] = typeName
}

// parseExprStmt parses expression statements. From the go/ast docs:
// 		An ExprStmt node represents a (stand-alone) expression in a statement list.
func parseExprStmt(in *ast.ExprStmt, nameToTypeMap map[string]string, out *set.Set) {
	switch c := in.X.(type) {
	case *ast.CallExpr:
		switch f := c.Fun.(type) {
		case *ast.Ident:
			out.Add(f.Name)
		case *ast.SelectorExpr:
			structVarName := f.X.(*ast.Ident).Name
			calledMethodName := f.Sel.Name
			if _, ok := nameToTypeMap[structVarName]; ok {
				out.Add(fmt.Sprintf("%s.%s", nameToTypeMap[structVarName], calledMethodName))
			}
		}
	}
}

// parseAssignStmt handles AssignStmt nodes. From the go/ast docs:
//    An AssignStmt node represents an assignment or a short variable declaration
func parseAssignStmt(in *ast.AssignStmt, nameToTypeMap map[string]string, helperFunctionReturnMap map[string][]string, out *set.Set) {
	leftHandSide := []string{}
	for i := range in.Lhs {
		switch v := in.Lhs[i].(type) {
		case *ast.Ident:
			varName := v.Name
			leftHandSide = append(leftHandSide, varName)
		}
	}

	for j := range in.Rhs {
		switch t := in.Rhs[j].(type) {
		case *ast.FuncLit:
			parseFuncLit(t, nameToTypeMap, helperFunctionReturnMap, out)
		case *ast.UnaryExpr:
			// FIXME: I think something might be goofy here, note the [0]
			parseUnaryExpr(t, leftHandSide[0], nameToTypeMap)
			//for x := range leftHandSide {
			//	parseUnaryExpr(t, leftHandSide[x], nameToTypeMap)
			//}
		case *ast.CallExpr:
			if len(in.Rhs) != len(in.Lhs) {
				var functionName string
				switch funcInfo := t.Fun.(type) {
				case *ast.Ident:
					functionName = funcInfo.Name
				case *ast.SelectorExpr:
					functionName = funcInfo.Sel.Name
				}
				if _, ok := helperFunctionReturnMap[functionName]; ok {
					for i, thing := range leftHandSide {
						nameToTypeMap[thing] = helperFunctionReturnMap[functionName][i]
					}
				}
			}
			parseCallExpr(t, nameToTypeMap, helperFunctionReturnMap, out)
		}
	}
}

// parseFuncLit parses a function literal. From the go/ast docs:
// 		A FuncLit node represents a function literal.
// FuncLits have bodies that we basically need to explore the same way that we explore a normal function.
func parseFuncLit(in *ast.FuncLit, nameToTypeMap map[string]string, helperFunctionReturnMap map[string][]string, out *set.Set) {
	for _, le := range in.Body.List {
		switch e := le.(type) {
		case *ast.AssignStmt: // handles things like `e := Example{}` (with or without &)
			parseAssignStmt(e, nameToTypeMap, helperFunctionReturnMap, out)
		case *ast.DeclStmt: // handles things like `var e Example`
			parseDeclStmt(e, nameToTypeMap)
		case *ast.ExprStmt: // handles function calls
			parseExprStmt(e, nameToTypeMap, out)
		}
	}
}

// parseStmt parses a statement. From the go/ast docs:
// 		 All statement nodes implement the Stmt interface.
func parseStmt(in ast.Stmt, nameToTypeMap map[string]string, helperFunctionReturnMap map[string][]string, out *set.Set) {
	switch e := in.(type) {
	case *ast.AssignStmt: // handles things like `e := Example{}` (with or without &)
		parseAssignStmt(e, nameToTypeMap, helperFunctionReturnMap, out)
	case *ast.RangeStmt:
		for _, x := range e.Body.List {
			parseStmt(x, nameToTypeMap, helperFunctionReturnMap, out)
		}
	case *ast.IfStmt:
		for _, x := range e.Body.List {
			parseStmt(x, nameToTypeMap, helperFunctionReturnMap, out)
		}
	case *ast.DeclStmt: // handles things like `var e Example`
		parseDeclStmt(e, nameToTypeMap)
	case *ast.ExprStmt: // handles function calls
		parseExprStmt(e, nameToTypeMap, out)
	}
}

func getDeclaredNames(in *ast.File, out *set.Set) {
	for _, d := range in.Decls {
		switch f := d.(type) {
		case *ast.FuncDecl:
			parseFuncDecl(f, out)
		}
	}
}

// parseFuncDecl parses function declarations. From the go/ast docs:
//		A FuncDecl node represents a function declaration.
func parseFuncDecl(f *ast.FuncDecl, out *set.Set) {
	functionName := f.Name.Name // "Avoid Stutter" lol
	var parentName string
	if f.Recv != nil {
		switch x := f.Recv.List[0].Type.(type) {
		case *ast.StarExpr:
			parentName = x.X.(*ast.Ident).Name
		case *ast.Ident:
			parentName = x.Obj.Name
		}
	}

	if parentName != "" {
		out.Add(fmt.Sprintf("%s.%s", parentName, functionName))
	} else {
		out.Add(functionName)
	}
}

func getCalledNames(in *ast.File, out *set.Set) {
	helperFunctionReturnMap := map[string][]string{}
	nameToTypeMap := map[string]string{}
	for _, d := range in.Decls {
		switch n := d.(type) {
		case *ast.GenDecl:
			parseGenDecl(n, nameToTypeMap)
		case *ast.FuncDecl:
			parseFuncDeclCall(n, nameToTypeMap, helperFunctionReturnMap, out)
		}
	}
}

// parseFuncDeclCall parses function declarations in the context of a function call.
func parseFuncDeclCall(in *ast.FuncDecl, nameToTypeMap map[string]string, helperFunctionReturnMap map[string][]string, out *set.Set) {
	functionName := in.Name.Name
	if !strings.HasPrefix(functionName, "Test") {
		if in.Type.Results != nil {
			for _, r := range in.Type.Results.List {
				switch rt := r.Type.(type) {
				case *ast.StarExpr:
					helperFunctionReturnMap[functionName] = append(helperFunctionReturnMap[functionName], rt.X.(*ast.Ident).Name)
				case *ast.Ident:
					helperFunctionReturnMap[functionName] = append(helperFunctionReturnMap[functionName], rt.Name)
				}
			}
		}
	}
	for _, le := range in.Body.List {
		parseStmt(le, nameToTypeMap, helperFunctionReturnMap, out)
	}
}

// parseGenDecl handles GenDecl nodes. From the go/ast docs:
//     A GenDecl node (generic declaration node) represents an import, constant, type or variable declaration.
func parseGenDecl(in *ast.GenDecl, nameToTypeMap map[string]string) {
	for _, spec := range in.Specs {
		switch global := spec.(type) {
		case *ast.ValueSpec: // for things like `var e Example` declared outside of functions
			varName := global.Names[0].Name
			if global.Type != nil {
				typeName := global.Type.(*ast.Ident).Name
				nameToTypeMap[varName] = typeName
			}
		}
	}
}
