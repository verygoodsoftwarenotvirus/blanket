package main

import (
	"go/parser"
	"go/ast"
	"go/token"
	"testing"

	"github.com/fatih/set"
	"github.com/stretchr/testify/assert"
)

func TestGetDeclaredNamesFromFile(t *testing.T) {
	t.Parallel()

	simple := func(t *testing.T) {
		in, err := parser.ParseFile(token.NewFileSet(), "example_packages/simple/main.go", nil, parser.AllErrors)
		if err != nil {
			t.Logf("failing because ParseFile returned error: %v", err)
			t.FailNow()
		}

		expectedDeclarations := []string{"A", "B", "C", "wrapper"}
		expected := set.New()
		for _, x := range expectedDeclarations {
			expected.Add(x)
		}

		actual := set.New()

		getDeclaredNamesFromFile(in, actual)

		assert.Equal(t, expected, actual, "expected output did not match actual output")
	}

	methods := func(t *testing.T) {
		in, err := parser.ParseFile(token.NewFileSet(), "example_packages/methods/main.go", nil, parser.AllErrors)
		if err != nil {
			t.Logf("failing because ParseFile returned error: %v", err)
			t.FailNow()
		}

		expectedDeclarations := []string{"Example.A", "Example.B", "Example.C", "wrapper"}
		expected := set.New()
		for _, x := range expectedDeclarations {
			expected.Add(x)
		}

		actual := set.New()
		getDeclaredNamesFromFile(in, actual)

		assert.Equal(t, expected, actual, "expected output did not match actual output")
	}

	subtests := []subtest{
		{
			Message: "simple package",
			Test:    simple,
		},
		{
			Message: "methods",
			Test:    methods,
		},
	}
	runSubtestSuite(t, subtests)
}

func TestGetCalledNames(t *testing.T) {
	simple := func(t *testing.T) {
		in, err := parser.ParseFile(token.NewFileSet(), "example_packages/simple/main_test.go", nil, parser.AllErrors)
		if err != nil {
			t.Logf("failing because ParseFile returned error: %v", err)
			t.FailNow()
		}

		expectedDeclarations := []string{"A", "C", "wrapper"}
		expected := set.New()
		for _, x := range expectedDeclarations {
			expected.Add(x)
		}

		actual := set.New()
		getCalledNames(in, actual)

		assert.Equal(t, expected, actual, "expected output did not match actual output")
	}

	methods := func(t *testing.T) {
		in, err := parser.ParseFile(token.NewFileSet(), "example_packages/methods/main_test.go", nil, parser.AllErrors)
		if err != nil {
			t.Logf("failing because ParseFile returned error: %v", err)
			t.FailNow()
		}

		expectedDeclarations := []string{"Example.A", "Example.C", "wrapper"}
		expected := set.New()
		for _, x := range expectedDeclarations {
			expected.Add(x)
		}

		actual := set.New()
		getCalledNames(in, actual)

		assert.Equal(t, expected, actual, "expected output did not match actual output")
	}

	subtests := []subtest{
		{
			Message: "simple package",
			Test:    simple,
		},
		{
			Message: "methods",
			Test:    methods,
		},
	}
	runSubtestSuite(t, subtests)
}

func TestParseCallExpr(t *testing.T) {
	t.Parallel()

	astIdentTest := func(t *testing.T) {
		exampleFunctionName := "function"

		input := &ast.CallExpr{
			Fun: &ast.Ident{Name: exampleFunctionName},
		}
		exampleNameToTypeMap := map[string]string{}
		exampleHelperFunctionMap := map[string][]string{}

		actual := set.New()
		expected := set.New(exampleFunctionName)

		parseCallExpr(input, exampleNameToTypeMap, exampleHelperFunctionMap, actual)

		assert.Equal(t, expected, actual, "expected function name to be added to output")
	}

	astSelectorExprTest := func(t *testing.T) {
		exampleVariableName := "instance"
		exampleCustomTypeName := "CustomType"

		input := &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X: &ast.Ident{Name: exampleVariableName},
				Sel: &ast.Ident{Name: "method"},
			},
		}
		exampleNameToTypeMap := map[string]string{
			exampleVariableName: exampleCustomTypeName,
		}
		exampleHelperFunctionMap := map[string][]string{}

		actual := set.New()
		expected := set.New("CustomType.method")

		parseCallExpr(input, exampleNameToTypeMap, exampleHelperFunctionMap, actual)

		assert.Equal(t, expected, actual, "expected function name to be added to output")
	}

	astSelectorExprTestWithoutMatchInMap := func(t *testing.T) {
		input := &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X: &ast.Ident{Name: "instance"},
				Sel: &ast.Ident{Name: "method"},
			},
		}
		exampleNameToTypeMap := map[string]string{}
		exampleHelperFunctionMap := map[string][]string{}

		actual := set.New()
		expected := set.New()

		parseCallExpr(input, exampleNameToTypeMap, exampleHelperFunctionMap, actual)

		assert.Equal(t, expected, actual, "expected function name to NOT be added to output")
	}

	t.Run("with ast.Ident", astIdentTest)
	t.Run("with ast.SelectorExpr", astSelectorExprTest)
	t.Run("with ast.SelectorExpr, but no matching entity", astSelectorExprTestWithoutMatchInMap)
}


func TestParseGenDecl(t *testing.T) {
	t.Parallel()

	actual := map[string]string{}
	exampleInput := &ast.GenDecl{
		Specs: []ast.Spec{
			&ast.ValueSpec{
				Type: &ast.Ident{Name: "type"},
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

func TestParseAssignStmt(t *testing.T) {
	input := &ast.AssignStmt{
		Lhs: []ast.Expr{
			&ast.Ident{
				Name: "x",
			},
		},
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.Ident{
					Name: "method",
				},
			},
		},
	}

	exampleNameToTypeMap := map[string]string{}
	exampleHelperFunctionMap := map[string][]string{}

	actual := set.New()
	expected := set.New("method")

	parseAssignStmt(input, exampleNameToTypeMap, exampleHelperFunctionMap, actual)

	assert.Equal(t, expected, actual, "wtf")
}
