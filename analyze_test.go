package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/fatih/set"
	"github.com/stretchr/testify/assert"
)

// func TestGetDeclaredNames(t *testing.T) {
// 	t.Parallel()

// 	simple := func(t *testing.T) {
// 		in, err := parser.ParseFile(token.NewFileSet(), "example_packages/simple/main.go", nil, parser.AllErrors)
// 		if err != nil {
// 			t.Logf("failing because ParseFile returned error: %v", err)
// 			t.FailNow()
// 		}

// 		expectedDeclarations := []string{"A", "B", "C", "wrapper"}
// 		expected := set.New()
// 		for _, x := range expectedDeclarations {
// 			expected.Add(x)
// 		}

// 		actual := set.New()

// 		getDeclaredNames(in, actual)

// 		assert.Equal(t, expected, actual, "expected output did not match actual output")
// 	}

// 	methods := func(t *testing.T) {
// 		in, err := parser.ParseFile(token.NewFileSet(), "example_packages/methods/main.go", nil, parser.AllErrors)
// 		if err != nil {
// 			t.Logf("failing because ParseFile returned error: %v", err)
// 			t.FailNow()
// 		}

// 		expectedDeclarations := []string{"Example.A", "Example.B", "Example.C", "wrapper"}
// 		expected := set.New()
// 		for _, x := range expectedDeclarations {
// 			expected.Add(x)
// 		}

// 		actual := set.New()
// 		getDeclaredNames(in, actual)

// 		assert.Equal(t, expected, actual, "expected output did not match actual output")
// 	}

// 	subtests := []subtest{
// 		{
// 			Message: "simple package",
// 			Test:    simple,
// 		},
// 		{
// 			Message: "methods",
// 			Test:    methods,
// 		},
// 	}
// 	runSubtestSuite(t, subtests)
// }

// func TestGetCalledNames(t *testing.T) {
// 	simple := func(t *testing.T) {
// 		in, err := parser.ParseFile(token.NewFileSet(), "example_packages/simple/main_test.go", nil, parser.AllErrors)
// 		if err != nil {
// 			t.Logf("failing because ParseFile returned error: %v", err)
// 			t.FailNow()
// 		}

// 		expectedDeclarations := []string{"A", "C", "wrapper"}
// 		expected := set.New()
// 		for _, x := range expectedDeclarations {
// 			expected.Add(x)
// 		}

// 		actual := set.New()
// 		getCalledNames(in, actual)

// 		assert.Equal(t, expected, actual, "expected output did not match actual output")
// 	}

// 	methods := func(t *testing.T) {
// 		in, err := parser.ParseFile(token.NewFileSet(), "example_packages/methods/main_test.go", nil, parser.AllErrors)
// 		if err != nil {
// 			t.Logf("failing because ParseFile returned error: %v", err)
// 			t.FailNow()
// 		}

// 		expectedDeclarations := []string{"Example.A", "Example.C", "wrapper"}
// 		expected := set.New()
// 		for _, x := range expectedDeclarations {
// 			expected.Add(x)
// 		}

// 		actual := set.New()
// 		getCalledNames(in, actual)

// 		assert.Equal(t, expected, actual, "expected output did not match actual output")
// 	}

// 	subtests := []subtest{
// 		{
// 			Message: "simple package",
// 			Test:    simple,
// 		},
// 		{
// 			Message: "methods",
// 			Test:    methods,
// 		},
// 	}
// 	runSubtestSuite(t, subtests)
// }

func TestParseCallExpr(t *testing.T) {
	t.Parallel()

	astIdentTest := func(t *testing.T) {
		codeSample := `
			package main
			var function func()
			func main(){
				fart := function()
			}
		`

		p, err := parser.ParseFile(token.NewFileSet(), "example.go", codeSample, parser.AllErrors)
		if err != nil {
			t.FailNow()
		}

		input := p.Decls[1].(*ast.FuncDecl).Body.List[0].(*ast.AssignStmt).Rhs[0].(*ast.CallExpr)
		exampleNameToTypeMap := map[string]string{}
		exampleHelperFunctionMap := map[string][]string{}

		actual := set.New()
		expected := set.New("function")

		parseCallExpr(input, exampleNameToTypeMap, exampleHelperFunctionMap, actual)

		assert.Equal(t, expected, actual, "expected function name to be added to output")
	}
	t.Run("with ast.Ident", astIdentTest)

	astSelectorExprTest := func(t *testing.T) {
		codeSample := `
			package main
			type Struct struct{}
			func (s Struct) method(){}
			func main(){
				s := Struct{}
				s.method()
			}
		`

		p, err := parser.ParseFile(token.NewFileSet(), "example.go", codeSample, parser.AllErrors)
		if err != nil {
			t.FailNow()
		}

		input := p.Decls[2].(*ast.FuncDecl).Body.List[1].(*ast.ExprStmt).X.(*ast.CallExpr)
		exampleNameToTypeMap := map[string]string{"s": "Struct"}
		exampleHelperFunctionMap := map[string][]string{}
		actual := set.New()
		expected := set.New("Struct.method")

		parseCallExpr(input, exampleNameToTypeMap, exampleHelperFunctionMap, actual)

		assert.Equal(t, expected, actual, "expected function name to be added to output")
	}
	t.Run("with ast.SelectorExpr", astSelectorExprTest)

	astSelectorExprTestWithoutMatchInMap := func(t *testing.T) {
		codeSample := `
			package main
			type Struct struct{}
			func (s Struct) method(){}
			func main(){
				s := Struct{}
				s.method()
			}
		`

		p, err := parser.ParseFile(token.NewFileSet(), "example.go", codeSample, parser.AllErrors)
		if err != nil {
			t.FailNow()
		}

		input := p.Decls[2].(*ast.FuncDecl).Body.List[1].(*ast.ExprStmt).X.(*ast.CallExpr)
		exampleNameToTypeMap := map[string]string{}
		exampleHelperFunctionMap := map[string][]string{}
		actual := set.New()
		expected := set.New()

		parseCallExpr(input, exampleNameToTypeMap, exampleHelperFunctionMap, actual)

		assert.Equal(t, expected, actual, "expected function name to be added to output")
	}
	t.Run("with ast.SelectorExpr, but no matching entit", astSelectorExprTestWithoutMatchInMap)
}

func TestParseUnaryExpr(t *testing.T) {
	t.Parallel()

	codeSample := `
			package main
			type Struct struct{}
			func main(){
				s := &Struct{}
			}
		`

	p, err := parser.ParseFile(token.NewFileSet(), "example.go", codeSample, parser.AllErrors)
	if err != nil {
		t.FailNow()
	}

	input := p.Decls[1].(*ast.FuncDecl).Body.List[0].(*ast.AssignStmt).Rhs[0].(*ast.UnaryExpr)
	expected := map[string]string{"s": "Struct"}
	actual := map[string]string{}

	parseUnaryExpr(input, "s", actual)

	assert.Equal(t, expected, actual, "actual output does not match expected output")
}

func TestParseDeclStmt(t *testing.T) {
	t.Parallel()

	codeSample := `
		package main
		func main(){
			var test bool
		}
	`

	p, err := parser.ParseFile(token.NewFileSet(), "example.go", codeSample, parser.AllErrors)
	if err != nil {
		t.FailNow()
	}

	input := p.Decls[0].(*ast.FuncDecl).Body.List[0].(*ast.DeclStmt)
	expected := map[string]string{"test": "bool"}
	actual := map[string]string{}

	parseDeclStmt(input, actual)

	assert.Equal(t, expected, actual, "actual output does not match expected output")
}

func TestParseExprStmt(t *testing.T) {
	t.Parallel()

	ident := func(t *testing.T) {
		exampleFunctionName := "function"
		exampleInput := &ast.ExprStmt{
			X: &ast.CallExpr{
				Fun: &ast.Ident{Name: exampleFunctionName},
			},
		}

		nameToTypeMap := map[string]string{}
		expected := set.New(exampleFunctionName)
		actual := set.New()

		parseExprStmt(exampleInput, nameToTypeMap, actual)

		assert.Equal(t, expected, actual, "actual output does not match expected output")
	}

	selector := func(t *testing.T) {
		exampleVarName := "var"
		exampleFunctionName := "method"
		exampleInput := &ast.ExprStmt{
			X: &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					Sel: &ast.Ident{Name: exampleFunctionName},
					X:   &ast.Ident{Name: exampleVarName},
				},
			},
		}

		nameToTypeMap := map[string]string{
			exampleVarName: "Example",
		}
		expected := set.New("Example.method")
		actual := set.New()

		parseExprStmt(exampleInput, nameToTypeMap, actual)

		assert.Equal(t, expected, actual, "actual output does not match expected output")
	}

	t.Run("CallExpr.Fun.(*ast.Ident)", ident)
	t.Run("CallExpr.Fun.(*ast.Selector)", selector)
}

func TestParseGenDecl(t *testing.T) {
	t.Parallel()

	actual := map[string]string{}
	exampleInput := &ast.GenDecl{
		Specs: []ast.Spec{
			&ast.ValueSpec{
				Type:  &ast.Ident{Name: "type"},
				Names: []*ast.Ident{{Name: "name"}},
			},
		},
	}
	expected := map[string]string{
		"name": "type",
	}

	parseGenDecl(exampleInput, actual)

	assert.Equal(t, expected, actual, "expected variable type and name to be inserted into map")
}

func TestParseFuncDecl(t *testing.T) {
	t.Parallel()

	simple := func(t *testing.T) {
		exampleFunctionName := "function"
		exampleInput := &ast.FuncDecl{
			Name: &ast.Ident{Name: exampleFunctionName},
		}

		expected := set.New(exampleFunctionName)
		actual := set.New()

		parseFuncDecl(exampleInput, actual)

		assert.Equal(t, expected, actual, "actual output does not match expected output")
	}

	methodASTIdentType := func(t *testing.T) {
		exampleStructName := "customObject"
		exampleFunctionName := "function"
		exampleInput := &ast.FuncDecl{
			Name: &ast.Ident{Name: exampleFunctionName},
			Recv: &ast.FieldList{
				List: []*ast.Field{
					{
						Type: &ast.Ident{Obj: &ast.Object{Name: exampleStructName}},
					},
				},
			},
		}

		expected := set.New(fmt.Sprintf("%s.%s", exampleStructName, exampleFunctionName))
		actual := set.New()

		parseFuncDecl(exampleInput, actual)

		assert.Equal(t, expected, actual, "actual output does not match expected output")
	}

	methodASTStarExprType := func(t *testing.T) {
		exampleStructName := "customObject"
		exampleFunctionName := "function"
		exampleInput := &ast.FuncDecl{
			Name: &ast.Ident{Name: exampleFunctionName},
			Recv: &ast.FieldList{
				List: []*ast.Field{
					{
						Type: &ast.StarExpr{
							X: &ast.Ident{Name: exampleStructName},
						},
					},
				},
			},
		}

		expected := set.New(fmt.Sprintf("%s.%s", exampleStructName, exampleFunctionName))
		actual := set.New()

		parseFuncDecl(exampleInput, actual)

		assert.Equal(t, expected, actual, "actual output does not match expected output")
	}

	t.Run("simple", simple)
	t.Run("with receiver", methodASTIdentType)
	t.Run("with ptr receiver", methodASTStarExprType)
}

func TestParseAssignStmt(t *testing.T) {
	t.Parallel()

	callExpr := func(t *testing.T) {
		exampleInput := &ast.AssignStmt{
			Lhs: []ast.Expr{
				&ast.Ident{Name: "x"},
			},
			Rhs: []ast.Expr{
				&ast.CallExpr{
					Fun: &ast.Ident{Name: "method"},
				},
			},
		}

		exampleNameToTypeMap := map[string]string{}
		exampleHelperFunctionMap := map[string][]string{}

		actual := set.New()
		expected := set.New("method")

		parseAssignStmt(exampleInput, exampleNameToTypeMap, exampleHelperFunctionMap, actual)

		assert.Equal(t, expected, actual, "actual output does not match expected output")
	}

	callExprWithMultipleReturnsAndIdent := func(t *testing.T) {
		exampleHelperFunctionName := "helperFunction"
		exampleInput := &ast.AssignStmt{
			Lhs: []ast.Expr{
				&ast.Ident{Name: "x"},
				&ast.Ident{Name: "y"},
			},
			Rhs: []ast.Expr{
				&ast.CallExpr{
					Fun: &ast.Ident{Name: exampleHelperFunctionName},
				},
			},
		}

		exampleHelperFunctionMap := map[string][]string{
			exampleHelperFunctionName: {
				"X",
				"Y",
			},
		}

		s := set.New()
		actual := map[string]string{}
		expected := map[string]string{
			"x": "X",
			"y": "Y",
		}

		parseAssignStmt(exampleInput, actual, exampleHelperFunctionMap, s)

		assert.Equal(t, expected, actual, "actual output does not match expected output")
	}

	callExprWithMultipleReturnsAndSelectorExpr := func(t *testing.T) {
		// FIXME: I'm not certain this test does what I think it should be doing.
		exampleHelperFunctionName := "helperFunction"
		exampleInput := &ast.AssignStmt{
			Lhs: []ast.Expr{
				&ast.Ident{Name: "x"},
				&ast.Ident{Name: "y"},
			},
			Rhs: []ast.Expr{
				&ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   &ast.Ident{Name: "name"},
						Sel: &ast.Ident{Name: exampleHelperFunctionName},
					},
				},
			},
		}

		exampleHelperFunctionMap := map[string][]string{
			exampleHelperFunctionName: {
				"X",
				"Y",
			},
		}

		out := set.New()
		actual := map[string]string{}
		expected := map[string]string{
			"x": "X",
			"y": "Y",
		}

		parseAssignStmt(exampleInput, actual, exampleHelperFunctionMap, out)

		assert.Equal(t, expected, actual, "actual output does not match expected output")
	}

	unaryExpr := func(t *testing.T) {
		// FIXME: I'm not certain this test does what I think it should be doing.
		exampleHelperFunctionName := "helperFunction"
		exampleExprName := "expression"
		exampleInput := &ast.AssignStmt{
			Lhs: []ast.Expr{
				&ast.Ident{Name: "x"},
				&ast.Ident{Name: "y"},
			},
			Rhs: []ast.Expr{
				&ast.UnaryExpr{
					X: &ast.CompositeLit{
						Type: &ast.Ident{Name: exampleExprName},
					},
				},
			},
		}

		exampleHelperFunctionMap := map[string][]string{
			exampleHelperFunctionName: {
				"X",
				"Y",
			},
		}

		out := set.New()
		actual := map[string]string{}
		expected := map[string]string{
			"x": "expression",
		}

		parseAssignStmt(exampleInput, actual, exampleHelperFunctionMap, out)

		assert.Equal(t, expected, actual, "actual output does not match expected output")
	}

	functionLiteral := func(t *testing.T) {
		// FIXME: I'm not certain this test does what I think it should be doing.
		exampleHelperFunctionName := "helperFunction"
		exampleInput := &ast.AssignStmt{
			Lhs: []ast.Expr{
				&ast.Ident{Name: "x"},
				&ast.Ident{Name: "y"},
			},
			Rhs: []ast.Expr{
				&ast.FuncLit{
					Body: &ast.BlockStmt{},
				},
			},
		}

		exampleHelperFunctionMap := map[string][]string{
			exampleHelperFunctionName: {
				"X",
				"Y",
			},
		}

		out := set.New()
		actual := map[string]string{}
		expected := map[string]string{}

		parseAssignStmt(exampleInput, actual, exampleHelperFunctionMap, out)

		assert.Equal(t, expected, actual, "actual output does not match expected output")
	}

	t.Run("CallExpr", callExpr)
	t.Run("CallExpr with multiple returns and ast.Ident Fun value", callExprWithMultipleReturnsAndIdent)
	t.Run("CallExpr with multiple returns and ast.SelectorExpr Fun value", callExprWithMultipleReturnsAndSelectorExpr)
	t.Run("UnaryExpr", unaryExpr)
	t.Run("FuncLit", functionLiteral)
}
