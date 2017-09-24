package main

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/fatih/set"
	"go/token"
)

func parseExpr(in ast.Expr, nameToTypeMap map[string]string, helperFunctionReturnMap map[string][]string, out *set.Set) {
	// FIXME: iterate over in.Args to see if there are function calls
	switch f := in.(type) {
	case *ast.Ident:
		functionName := f.Name
		if _, ok := helperFunctionReturnMap[functionName]; !ok {
			out.Add(functionName)
		}
	case *ast.SelectorExpr:
		if x, ok := f.X.(*ast.Ident); ok {
			structVarName := x.Name
			calledMethodName := f.Sel.Name
			if _, ok := nameToTypeMap[structVarName]; ok {
				out.Add(fmt.Sprintf("%s.%s", nameToTypeMap[structVarName], calledMethodName))
			}
		}
	case *ast.FuncLit:
		parseFuncLit(f, nameToTypeMap, helperFunctionReturnMap, out)
	}
}

func parseCallExpr(in *ast.CallExpr, nameToTypeMap map[string]string, helperFunctionReturnMap map[string][]string, out *set.Set) {
	for _, a := range in.Args {
		if r, ok := a.(*ast.CallExpr); ok {
			parseCallExpr(r, nameToTypeMap, helperFunctionReturnMap, out)
		}
	}
	parseExpr(in.Fun, nameToTypeMap, helperFunctionReturnMap, out)
}

// parseUnaryExpr parses Unary expressions. From the go/ast docs:
//      A UnaryExpr node represents a unary expression. Unary "*" expressions are represented via StarExpr nodes.
// (handles declarations like `callExpr := &ast.UnaryExpr{}` or `callExpr := ast.UnaryExpr{}`)
func parseUnaryExpr(in *ast.UnaryExpr, varName string, nameToTypeMap map[string]string, helperFunctionReturnMap map[string][]string, out *set.Set) {
	if cl, ok := in.X.(*ast.CompositeLit); ok {
		for _, e := range cl.Elts {
			switch et := e.(type) {
			case *ast.CallExpr:
				parseExpr(et.Fun, nameToTypeMap, helperFunctionReturnMap, out)
			case *ast.KeyValueExpr:
				switch vt := et.Value.(type) {
				case *ast.CallExpr:
					parseCallExpr(vt, nameToTypeMap, helperFunctionReturnMap, out)
				}
			}
		}
		switch u := cl.Type.(type) {
		case *ast.Ident:
			nameToTypeMap[varName] = u.Name
		case *ast.SelectorExpr:
			nameToTypeMap[varName] = u.Sel.Name
		}
	}
}

// parseDeclStmt parses declaration statments. From the go/ast docs:
// 		A DeclStmt node represents a declaration in a statement list.
// DeclStmts come from function bodies, GenDecls come from package-wide const or var declarations
func parseDeclStmt(in *ast.DeclStmt, nameToTypeMap map[string]string) {
	// FIXME: we make a whole mess of assumptions right here. I haven't thusfar seen any
	// 		  evidence that these assumptions are incorrect or dangerous, but that doesn't
	// 		  mean they don't carry the inherent risk that most assumptions do.
	if s, ok := in.Decl.(*ast.GenDecl).Specs[0].(*ast.ValueSpec); ok {
		varName := s.Names[0].Name
		switch t := s.Type.(type) {
		case *ast.Ident:
			nameToTypeMap[varName] = t.Name
		case *ast.SelectorExpr:
			nameToTypeMap[varName] = t.Sel.Name
		}
	}
}

// parseExprStmt parses expression statements. From the go/ast docs:
// 		An ExprStmt node represents a (stand-alone) expression in a statement list.
func parseExprStmt(in *ast.ExprStmt, nameToTypeMap map[string]string, helperFunctionReturnMap map[string][]string, out *set.Set) {
	if c, ok := in.X.(*ast.CallExpr); ok {
		parseCallExpr(c, nameToTypeMap, helperFunctionReturnMap, out)
	}
}

func parseCompositeLit(in *ast.CompositeLit, varName string, nameToTypeMap map[string]string, helperFunctionReturnMap map[string][]string, out *set.Set) {
	for _, e := range in.Elts {
		switch et := e.(type) {
		case *ast.CallExpr:
			parseExpr(et.Fun, nameToTypeMap, helperFunctionReturnMap, out)
		}
	}

	switch t := in.Type.(type) {
	case *ast.Ident:
		nameToTypeMap[varName] = t.Name
	case *ast.SelectorExpr:
		nameToTypeMap[varName] = t.Sel.Name
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
			parseUnaryExpr(t, leftHandSide[j], nameToTypeMap, helperFunctionReturnMap, out)
		case *ast.CompositeLit:
			if len(leftHandSide) > j {
				parseCompositeLit(t, leftHandSide[j], nameToTypeMap, helperFunctionReturnMap, out)
			} else {
				parseCompositeLit(t, "", nameToTypeMap, helperFunctionReturnMap, out)
			}
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

func parseHelperFunction(in *ast.FuncDecl, helperFunctionReturnMap map[string][]string, out *set.Set) {
	functionName := in.Name.Name
	if in.Type.Results != nil {
		for _, r := range in.Type.Results.List {
			switch rt := r.Type.(type) {
			case *ast.StarExpr:
				switch x := rt.X.(type) {
				case *ast.Ident:
					helperFunctionReturnMap[functionName] = append(helperFunctionReturnMap[functionName], x.Name)
				case *ast.SelectorExpr:
					if pkg, ok := x.X.(*ast.Ident); ok {
						pkgName := pkg.Name
						pkgStruct := x.Sel.Name
						helperFunctionReturnMap[functionName] = append(helperFunctionReturnMap[functionName], fmt.Sprintf("%s.%s", pkgName, pkgStruct))
					}
				}
			case *ast.Ident:
				helperFunctionReturnMap[functionName] = append(helperFunctionReturnMap[functionName], rt.Name)
			}
		}
	}
}

// parseTestFuncDecl parses function declarations that don't fit the testing function shape. The purpose of this is to catch
// functions that authors may build to generate certain values for tests.
func parseTestFuncDecl(in *ast.FuncDecl, nameToTypeMap map[string]string, helperFunctionReturnMap map[string][]string, out *set.Set) {
	functionName := in.Name.Name
	if !strings.HasPrefix(functionName, "Test") {
		parseHelperFunction(in, helperFunctionReturnMap, out)
	}
	if in.Body != nil {
		for _, le := range in.Body.List {
			parseStmt(le, nameToTypeMap, helperFunctionReturnMap, out)
		}
	}
}

// parseFuncLit parses a function literal. From the go/ast docs:
// 		A FuncLit node represents a function literal.
// FuncLits have bodies that we basically need to explore the same way that we explore a normal function.
func parseFuncLit(in *ast.FuncLit, nameToTypeMap map[string]string, helperFunctionReturnMap map[string][]string, out *set.Set) {
	for _, le := range in.Body.List {
		parseStmt(le, nameToTypeMap, helperFunctionReturnMap, out)
	}
}

func parseReturnStmt(in *ast.ReturnStmt, nameToTypeMap map[string]string, helperFunctionReturnMap map[string][]string, out *set.Set) {
	for _, x := range in.Results {
		switch y := x.(type) {
		case *ast.CallExpr:
			parseExpr(y.Fun, nameToTypeMap, helperFunctionReturnMap, out)
		}
	}
}

func parseSelectStmt(in *ast.SelectStmt, nameToTypeMap map[string]string, helperFunctionReturnMap map[string][]string, out *set.Set) {
	for _, x := range in.Body.List {
		switch y := x.(type) {
		case *ast.CommClause:
			for _, z := range y.Body {
				parseStmt(z, nameToTypeMap, helperFunctionReturnMap, out)
			}
		}
	}
}

// parseSendStmt parses a send statement. (<-)
func parseSendStmt(in *ast.SendStmt, nameToTypeMap map[string]string, helperFunctionReturnMap map[string][]string, out *set.Set) {
	switch n := in.Value.(type) {
	case *ast.CallExpr:
		parseCallExpr(n, nameToTypeMap, helperFunctionReturnMap, out)
	}
}

func parseSwitchStmt(in *ast.SwitchStmt, nameToTypeMap map[string]string, helperFunctionReturnMap map[string][]string, out *set.Set) {
	for _, x := range in.Body.List {
		switch y := x.(type) {
		case *ast.CaseClause:
			for _, z := range y.Body {
				parseStmt(z, nameToTypeMap, helperFunctionReturnMap, out)
			}
		}
	}
}

// parseTypeSwitchStmt parses
func parseTypeSwitchStmt(in *ast.TypeSwitchStmt, nameToTypeMap map[string]string, helperFunctionReturnMap map[string][]string, out *set.Set) {
	for _, x := range in.Body.List {
		switch y := x.(type) {
		case *ast.CaseClause:
			for _, z := range y.Body {
				parseStmt(z, nameToTypeMap, helperFunctionReturnMap, out)
			}
		}
	}
}

// parseStmt parses a statement. From the go/ast docs:
// 		All statement nodes implement the Stmt interface.
// Cases we don't handle:
//		BadStmt - we only parse valid code
//		BlockStmt (sort of, we iterate over these in the form of `x.Body.List`)
//		these are simply unnecessary:
//			BranchStmt
//			EmptyStmt
//			IncDeclStmt
//			LabeledStmt
func parseStmt(in ast.Stmt, nameToTypeMap map[string]string, helperFunctionReturnMap map[string][]string, out *set.Set) {
	switch e := in.(type) {
	case *ast.AssignStmt: // handles things like `e := Example{}` (with or without &)
		parseAssignStmt(e, nameToTypeMap, helperFunctionReturnMap, out)
	// NOTE: even though RangeStmt/IfStmt/ForStmt are handled identically, Go will (rightfully) complain when trying
	// to use a multiple case statement (i.e. `case *ast.RangeStmt, *ast.IfStmt`), so we're doing it this way.
	case *ast.RangeStmt:
		for _, x := range e.Body.List {
			parseStmt(x, nameToTypeMap, helperFunctionReturnMap, out)
		}
	case *ast.IfStmt:
		for _, x := range e.Body.List {
			parseStmt(x, nameToTypeMap, helperFunctionReturnMap, out)
		}
	case *ast.ForStmt:
		for _, x := range e.Body.List {
			parseStmt(x, nameToTypeMap, helperFunctionReturnMap, out)
		}
	case *ast.DeclStmt:
		parseDeclStmt(e, nameToTypeMap)
	case *ast.ExprStmt:
		parseExprStmt(e, nameToTypeMap, helperFunctionReturnMap, out)
	case *ast.DeferStmt:
		parseExpr(e.Call.Fun, nameToTypeMap, helperFunctionReturnMap, out)
	case *ast.GoStmt:
		parseExpr(e.Call.Fun, nameToTypeMap, helperFunctionReturnMap, out)
	case *ast.ReturnStmt:
		parseReturnStmt(e, nameToTypeMap, helperFunctionReturnMap, out)
	case *ast.SelectStmt:
		parseSelectStmt(e, nameToTypeMap, helperFunctionReturnMap, out)
	case *ast.SendStmt:
		parseSendStmt(e, nameToTypeMap, helperFunctionReturnMap, out)
	case *ast.SwitchStmt:
		parseSwitchStmt(e, nameToTypeMap, helperFunctionReturnMap, out)
	case *ast.TypeSwitchStmt:
		parseTypeSwitchStmt(e, nameToTypeMap, helperFunctionReturnMap, out)
	}
}

func getDeclaredNames(in *ast.File, fileset *token.FileSet, declaredFuncDetails map[string]TarpFunc) {
	for _, d := range in.Decls {
		switch f := d.(type) {
		case *ast.FuncDecl:
			declPos := fileset.Position(f.Type.Func)
			functionName := parseFuncDecl(f)

			tf := TarpFunc{
				Name:     functionName,
				Filename: declPos.Filename,
				DeclPos:  declPos,
			}

			if f.Body != nil {
				tf.RBracePos = fileset.Position(f.Body.Lbrace)
				tf.LBracePos = fileset.Position(f.Body.Rbrace)
			}
			declaredFuncDetails[functionName] = tf
		}
	}
}

// parseFuncDecl parses function declarations. From the go/ast docs:
//		A FuncDecl node represents a function declaration.
func parseFuncDecl(f *ast.FuncDecl) string {
	functionName := f.Name.Name // "Avoid Stutter" lol
	var parentName string
	if f.Recv != nil {
		switch x := f.Recv.List[0].Type.(type) {
		case *ast.StarExpr:
			if parent, ok := x.X.(*ast.Ident); ok {
				parentName = parent.Name
			}
		case *ast.Ident:
			parentName = x.Obj.Name
		}
	}

	if parentName != "" {
		return fmt.Sprintf("%s.%s", parentName, functionName)
	}
	return functionName

}

func getCalledNames(in *ast.File, nameToTypeMap map[string]string, helperFunctionReturnMap map[string][]string, out *set.Set) {
	for _, d := range in.Decls {
		switch n := d.(type) {
		case *ast.GenDecl:
			parseGenDecl(n, nameToTypeMap)
		case *ast.FuncDecl:
			parseTestFuncDecl(n, nameToTypeMap, helperFunctionReturnMap, out)
		}
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
				if t, ok := global.Type.(*ast.Ident); ok {
					typeName := t.Name
					nameToTypeMap[varName] = typeName
				}
			}
		}
	}
}
