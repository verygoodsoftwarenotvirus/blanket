package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/fatih/set"
	"github.com/stretchr/testify/assert"
)

func parseChunkOfCode(t *testing.T, chunkOfCode string) *ast.File {
	p, err := parser.ParseFile(token.NewFileSet(), "example.go", chunkOfCode, parser.AllErrors)
	if err != nil {
		t.FailNow()
	}
	return p
}

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

		p := parseChunkOfCode(t, codeSample)
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

		p := parseChunkOfCode(t, codeSample)
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

		p := parseChunkOfCode(t, codeSample)
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

	p := parseChunkOfCode(t, codeSample)
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

	p := parseChunkOfCode(t, codeSample)
	input := p.Decls[0].(*ast.FuncDecl).Body.List[0].(*ast.DeclStmt)
	expected := map[string]string{"test": "bool"}
	actual := map[string]string{}

	parseDeclStmt(input, actual)

	assert.Equal(t, expected, actual, "actual output does not match expected output")
}

func TestParseExprStmt(t *testing.T) {
	t.Parallel()

	ident := func(t *testing.T) {
		codeSample := `
			package main
			var example func()
			func main(){
				example()
			}
		`
		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[1].(*ast.FuncDecl).Body.List[0].(*ast.ExprStmt)

		nameToTypeMap := map[string]string{}
		expected := set.New("example")
		actual := set.New()

		parseExprStmt(input, nameToTypeMap, actual)

		assert.Equal(t, expected, actual, "actual output does not match expected output")
	}
	t.Run("CallExpr.Fun.(*ast.Ident)", ident)

	selector := func(t *testing.T) {
		codeSample := `
			package main
			type Example struct{}
			func (e Example) method() {}
			func main() {
				var e Example
				e.method()
			}

		`
		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[2].(*ast.FuncDecl).Body.List[1].(*ast.ExprStmt)

		nameToTypeMap := map[string]string{"e": "Example"}
		expected := set.New("Example.method")
		actual := set.New()

		parseExprStmt(input, nameToTypeMap, actual)

		assert.Equal(t, expected, actual, "actual output does not match expected output")
	}

	t.Run("CallExpr.Fun.(*ast.Selector)", selector)
}

func TestParseGenDecl(t *testing.T) {
	t.Parallel()

	codeSample := `
		package main
		var thing string
		func main(){}
	`

	p := parseChunkOfCode(t, codeSample)
	input := p.Decls[0].(*ast.GenDecl)
	actual := map[string]string{}
	expected := map[string]string{"thing": "string"}

	parseGenDecl(input, actual)

	assert.Equal(t, expected, actual, "expected function name to be added to output")
}

func TestParseFuncDecl(t *testing.T) {
	t.Parallel()

	simple := func(t *testing.T) {
		codeSample := `
			package test
			func example(){}
		`

		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[0].(*ast.FuncDecl)

		expected := set.New("example")
		actual := set.New()

		parseFuncDecl(input, actual)

		assert.Equal(t, expected, actual, "actual output does not match expected output")
	}
	t.Run("simple", simple)

	methodASTIdentType := func(t *testing.T) {
		codeSample := `
			package test
			type Example struct{}
			func (e Example) method(){}
		`

		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[1].(*ast.FuncDecl)

		expected := set.New("Example.method")
		actual := set.New()

		parseFuncDecl(input, actual)

		assert.Equal(t, expected, actual, "actual output does not match expected output")
	}
	t.Run("with receiver", methodASTIdentType)

	methodASTStarExprType := func(t *testing.T) {
		codeSample := `
			package test
			type Example struct{}
			func (e *Example) method(){}
		`

		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[1].(*ast.FuncDecl)

		expected := set.New("Example.method")
		actual := set.New()

		parseFuncDecl(input, actual)

		assert.Equal(t, expected, actual, "actual output does not match expected output")
	}
	t.Run("with ptr receiver", methodASTStarExprType)
}

func TestParseAssignStmt(t *testing.T) {
	t.Parallel()

	callExpr := func(t *testing.T) {
		codeSample := `
			package main
			import "testing"
			func example() error {
				return nil
			}
			func test() {
				e := example()
			}
		`

		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[2].(*ast.FuncDecl).Body.List[0].(*ast.AssignStmt)

		exampleNameToTypeMap := map[string]string{}
		exampleHelperFunctionMap := map[string][]string{}

		actual := set.New()
		expected := set.New("example")

		parseAssignStmt(input, exampleNameToTypeMap, exampleHelperFunctionMap, actual)

		assert.Equal(t, expected, actual, "actual output does not match expected output")
	}
	t.Run("CallExpr", callExpr)

	callExprWithMultipleReturnsAndIdent := func(t *testing.T) {
		// this case handles when a helper function is declared in another file.
		codeSample := `
			package main
			import "testing"
			func TestX() {
				x, y := example()
			}
		`

		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[1].(*ast.FuncDecl).Body.List[0].(*ast.AssignStmt)

		s := set.New()
		actual := map[string]string{}
		expected := map[string]string{
			"x": "X",
			"y": "Y",
		}
		exampleHelperFunctionMap := map[string][]string{
			"example": {
				"X",
				"Y",
			},
		}

		parseAssignStmt(input, actual, exampleHelperFunctionMap, s)

		assert.Equal(t, expected, actual, "actual output does not match expected output")
	}
	t.Run("CallExpr with multiple returns and ast.Ident Fun value", callExprWithMultipleReturnsAndIdent)

	callExprWithUnfamiliarSelectorExprAndMultipleReturn := func(t *testing.T) {
		codeSample := `
			package main
			import "testing"
			func TestX() {
				req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
			}
		`

		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[1].(*ast.FuncDecl).Body.List[0].(*ast.AssignStmt)

		s := set.New()
		actual := map[string]string{}
		expected := map[string]string{}
		exampleHelperFunctionMap := map[string][]string{}

		parseAssignStmt(input, actual, exampleHelperFunctionMap, s)

		assert.Equal(t, expected, actual, "actual output does not match expected output")
	}

	t.Run("Assign statement with multiple returns from some external function", callExprWithUnfamiliarSelectorExprAndMultipleReturn)

	callExprWithKnownSelectorExprAndMultipleReturn := func(t *testing.T) {
		codeSample := `
			package main
			import "testing"
			func TestX() {
				req, err := someHelperFunctionForTestsOnly()
			}
		`

		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[1].(*ast.FuncDecl).Body.List[0].(*ast.AssignStmt)

		s := set.New()
		actual := map[string]string{}
		expected := map[string]string{
			"req": "http.Request",
			"err": "error",
		}
		exampleHelperFunctionMap := map[string][]string{
			"someHelperFunctionForTestsOnly": {
				"http.Request",
				"error",
			},
		}

		parseAssignStmt(input, actual, exampleHelperFunctionMap, s)

		assert.Equal(t, expected, actual, "actual output does not match expected output")
	}
	t.Run("Assign statement with multiple returns from some internal function", callExprWithKnownSelectorExprAndMultipleReturn)

	unaryExpr := func(t *testing.T) {
		codeSample := `
			package main
			import "testing"
			func TestX(t *testing.T) {
				test := &SomeStruct{}
			}
		`

		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[1].(*ast.FuncDecl).Body.List[0].(*ast.AssignStmt)

		exampleHelperFunctionMap := map[string][]string{}

		out := set.New()
		actual := map[string]string{}
		expected := map[string]string{
			"test": "SomeStruct",
		}

		parseAssignStmt(input, actual, exampleHelperFunctionMap, out)

		assert.Equal(t, expected, actual, "actual output does not match expected output")
	}
	t.Run("UnaryExpr", unaryExpr)

	multipleUnaryExpressions := func(t *testing.T) {
		codeSample := `
			package main
			import "testing"
			func TestX(t *testing.T) {
				one, other := &SomeStruct{}, &SomeOtherStruct{}
			}
		`

		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[1].(*ast.FuncDecl).Body.List[0].(*ast.AssignStmt)

		exampleHelperFunctionMap := map[string][]string{}

		out := set.New()
		actual := map[string]string{}
		expected := map[string]string{
			"one": "SomeStruct",
			"other": "SomeOtherStruct",
		}

		parseAssignStmt(input, actual, exampleHelperFunctionMap, out)

		assert.Equal(t, expected, actual, "actual output does not match expected output")
	}
	t.Run("multiple unary expressions", multipleUnaryExpressions)

	functionLiteral := func(t *testing.T) {
		codeSample := `
		 	package main
		 	import "testing"
		 	func TestX(t *testing. T) {
		 		subtest := func(t *testing.T) {}
		 		t.Run("subtest", subtest)
		 	}
		 `

		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[1].(*ast.FuncDecl).Body.List[0].(*ast.AssignStmt)

		out := set.New()
		exampleHelperFunctionMap := map[string][]string{}
		actual := map[string]string{}
		expected := map[string]string{}

		parseAssignStmt(input, actual, exampleHelperFunctionMap, out)

		assert.Equal(t, expected, actual, "actual output does not match expected output")
	}
	t.Run("FuncLit", functionLiteral)
}
