package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"testing"

	"github.com/fatih/set"
	"github.com/stretchr/testify/assert"
)

////////////////////////////////////////////////////////
//                                                    //
//               Test Helper Functions                //
//                                                    //
////////////////////////////////////////////////////////

func parseChunkOfCode(t *testing.T, chunkOfCode string) *ast.File {
	p, err := parser.ParseFile(token.NewFileSet(), "example.go", chunkOfCode, parser.AllErrors)
	if err != nil {
		log.Println(err)
		t.FailNow()
	}
	return p
}

////////////////////////////////////////////////////////
//                                                    //
//                   Actual Tests                     //
//                                                    //
////////////////////////////////////////////////////////

func TestParseExpr(t *testing.T) {
	t.Run("ident", func(t *testing.T) {
		codeSample := `
			package main

			func main() {
				functionCall()
			}
		`

		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[0].(*ast.FuncDecl).Body.List[0].(*ast.ExprStmt).X.(*ast.CallExpr).Fun

		actual := set.New()
		expected := set.New("functionCall")

		parseExpr(input, map[string]string{}, map[string][]string{}, actual)

		assert.Equal(t, expected, actual, "expected function name to be added to output")
	})

	t.Run("selector", func(t *testing.T) {
		codeSample := `
			package main

			func main() {
				class.methodCall()
			}
		`

		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[0].(*ast.FuncDecl).Body.List[0].(*ast.ExprStmt).X.(*ast.CallExpr).Fun
		nameToTypeMap := map[string]string{"class": "Example"}

		actual := set.New()
		expected := set.New("Example.methodCall")

		parseExpr(input, nameToTypeMap, map[string][]string{}, actual)

		assert.Equal(t, expected, actual, "expected function name to be added to output")
	})

	t.Run("function literal", func(t *testing.T) {
		codeSample := `
			package main

			func main() {
				func() {
					functionCall()
				}()
			}
		`

		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[0].(*ast.FuncDecl).Body.List[0].(*ast.ExprStmt).X.(*ast.CallExpr).Fun

		actual := set.New()
		expected := set.New("functionCall")

		parseExpr(input, map[string]string{}, map[string][]string{}, actual)

		assert.Equal(t, expected, actual, "expected function name to be added to output")
	})
}

func TestParseCallExpr(t *testing.T) {
	t.Run("with ast.Ident", func(t *testing.T) {
		codeSample := `
			package main
			var function func()
			func main(){
				fart := function()
			}
		`

		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[1].(*ast.FuncDecl).Body.List[0].(*ast.AssignStmt).Rhs[0].(*ast.CallExpr)

		actual := set.New()
		expected := set.New("function")

		parseCallExpr(input, map[string]string{}, map[string][]string{}, actual)

		assert.Equal(t, expected, actual, "expected function name to be added to output")
	})

	t.Run("with ast.SelectorExpr", func(t *testing.T) {
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

		actual := set.New()
		expected := set.New("Struct.method")

		parseCallExpr(input, map[string]string{"s": "Struct"}, map[string][]string{}, actual)

		assert.Equal(t, expected, actual, "expected function name to be added to output")
	})

	t.Run("with ast.SelectorExpr, but no matching entity", func(t *testing.T) {
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

		actual := set.New()
		expected := set.New()

		parseCallExpr(input, map[string]string{}, map[string][]string{}, actual)

		assert.Equal(t, expected, actual, "expected function name to be added to output")
	})

	t.Run("with funcLit in argument list", func(_t *testing.T) {
		_t.Parallel()

		codeSample := `
			package main

			import "log"

			func main(){
				arbitraryCallExpression(func(i int) {
					log.Println(i)
				})
			}
		`

		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[1].(*ast.FuncDecl).Body.List[0].(*ast.ExprStmt).X.(*ast.CallExpr)

		actual := set.New()
		expected := set.New("arbitraryCallExpression")

		parseCallExpr(input, map[string]string{}, map[string][]string{}, actual)

		assert.Equal(t, expected, actual, "expected function name to be added to output")
	})
}

func TestParseUnaryExpr(t *testing.T) {
	codeSample := `
			package main
			type Struct struct{}
			func main(){
				s := &Struct{}
			}
		`

	p := parseChunkOfCode(t, codeSample)
	input := p.Decls[1].(*ast.FuncDecl).Body.List[0].(*ast.AssignStmt).Rhs[0].(*ast.UnaryExpr)

	actual := map[string]string{}
	expected := map[string]string{"s": "Struct"}

	parseUnaryExpr(input, "s", actual, map[string][]string{}, set.New())

	assert.Equal(t, expected, actual, "actual output does not match expected output")
}

func TestParseDeclStmt(t *testing.T) {
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
	t.Run("CallExpr.Fun.(*ast.Ident)", func(t *testing.T) {
		codeSample := `
			package main
			var example func()
			func main(){
				example()
			}
		`
		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[1].(*ast.FuncDecl).Body.List[0].(*ast.ExprStmt)

		expected := set.New("example")
		actual := set.New()

		parseExprStmt(input, map[string]string{}, map[string][]string{}, actual)

		assert.Equal(t, expected, actual, "actual output does not match expected output")
	})

	t.Run("CallExpr.Fun.(*ast.Selector)", func(t *testing.T) {
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

		expected := set.New("Example.method")
		actual := set.New()

		parseExprStmt(input, map[string]string{"e": "Example"}, map[string][]string{}, actual)

		assert.Equal(t, expected, actual, "actual output does not match expected output")
	})
}

func TestParseCompositeLit(t *testing.T) {
	t.Run("ident", func(t *testing.T) {
		codeSample := `
			package main
			func main() {
				x := &Example{
					methodCallAsArg(),
				}
			}

		`
		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[0].(*ast.FuncDecl).Body.List[0].(*ast.AssignStmt).Rhs[0].(*ast.UnaryExpr).X.(*ast.CompositeLit)

		expected := set.New("methodCallAsArg")
		actual := set.New()

		parseCompositeLit(input, "e", map[string]string{"e": "Example"}, map[string][]string{}, actual)

		assert.Equal(t, expected, actual, "actual output does not match expected output")
	})

	t.Run("selector", func(t *testing.T) {
		codeSample := `
			package main
			func main() {
				x := &pkg.Example{
					e.methodCallAsArg(),
				}
			}

		`
		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[0].(*ast.FuncDecl).Body.List[0].(*ast.AssignStmt).Rhs[0].(*ast.UnaryExpr).X.(*ast.CompositeLit)

		expected := set.New("Example.methodCallAsArg")
		actual := set.New()

		parseCompositeLit(input, "e", map[string]string{"e": "Example"}, map[string][]string{}, actual)

		assert.Equal(t, expected, actual, "actual output does not match expected output")
	})
}

func TestParseGenDecl(t *testing.T) {
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
	t.Run("simple", func(t *testing.T) {
		codeSample := `
			package test
			func example(){}
		`

		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[0].(*ast.FuncDecl)

		expected := "example"
		actual := parseFuncDecl(input)

		assert.Equal(t, expected, actual, "actual output does not match expected output")
	})

	t.Run("with receiver", func(t *testing.T) {
		codeSample := `
			package test
			type Example struct{}
			func (e Example) method(){}
		`

		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[1].(*ast.FuncDecl)

		expected := "Example.method"
		actual := parseFuncDecl(input)

		assert.Equal(t, expected, actual, "actual output does not match expected output")
	})

	t.Run("with ptr receiver", func(t *testing.T) {
		codeSample := `
			package test
			type Example struct{}
			func (e *Example) method(){}
		`

		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[1].(*ast.FuncDecl)

		expected := "Example.method"
		actual := parseFuncDecl(input)

		assert.Equal(t, expected, actual, "actual output does not match expected output")
	})
}

func TestParseAssignStmt(t *testing.T) {
	t.Run("CallExpr", func(t *testing.T) {
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

		actual := set.New()
		expected := set.New("example")

		parseAssignStmt(input, map[string]string{}, map[string][]string{}, actual)

		assert.Equal(t, expected, actual, "actual output does not match expected output")
	})

	t.Run("CallExpr with multiple returns and ast.Ident Fun value", func(t *testing.T) {
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

		actual := map[string]string{}
		expected := map[string]string{
			"x": "X",
			"y": "Y",
		}

		parseAssignStmt(input, actual, map[string][]string{"example": {"X", "Y"}}, set.New())

		assert.Equal(t, expected, actual, "actual output does not match expected output")
	})

	t.Run("Assign statement with multiple returns from some external function", func(t *testing.T) {
		codeSample := `
			package main
			import "testing"
			func TestX() {
				req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
			}
		`

		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[1].(*ast.FuncDecl).Body.List[0].(*ast.AssignStmt)

		actual := map[string]string{}
		expected := map[string]string{}

		parseAssignStmt(input, actual, map[string][]string{}, set.New())

		assert.Equal(t, expected, actual, "actual output does not match expected output")
	})

	t.Run("Assign statement with multiple returns from some internal function", func(t *testing.T) {
		codeSample := `
			package main
			import "testing"
			func TestX() {
				req, err := someHelperFunctionForTestsOnly()
			}
		`

		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[1].(*ast.FuncDecl).Body.List[0].(*ast.AssignStmt)

		actual := map[string]string{}
		expected := map[string]string{
			"req": "http.Request",
			"err": "error",
		}

		parseAssignStmt(input, actual, map[string][]string{"someHelperFunctionForTestsOnly": {"http.Request", "error"}}, set.New())

		assert.Equal(t, expected, actual, "actual output does not match expected output")
	})

	t.Run("UnaryExpr", func(t *testing.T) {
		codeSample := `
			package main
			import "testing"
			func TestX(t *testing.T) {
				test := &SomeStruct{}
			}
		`

		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[1].(*ast.FuncDecl).Body.List[0].(*ast.AssignStmt)

		actual := map[string]string{}
		expected := map[string]string{
			"test": "SomeStruct",
		}

		parseAssignStmt(input, actual, map[string][]string{}, set.New())

		assert.Equal(t, expected, actual, "actual output does not match expected output")
	})

	t.Run("multiple unary expressions", func(t *testing.T) {
		codeSample := `
			package main
			import "testing"
			func TestX(t *testing.T) {
				one, other := &SomeStruct{}, &SomeOtherStruct{}
			}
		`

		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[1].(*ast.FuncDecl).Body.List[0].(*ast.AssignStmt)

		actual := map[string]string{}
		expected := map[string]string{
			"one":   "SomeStruct",
			"other": "SomeOtherStruct",
		}

		parseAssignStmt(input, actual, map[string][]string{}, set.New())

		assert.Equal(t, expected, actual, "actual output does not match expected output")
	})

	t.Run("FuncLit", func(t *testing.T) {
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

		actual := map[string]string{}
		expected := map[string]string{}

		parseAssignStmt(input, actual, map[string][]string{}, set.New())

		assert.Equal(t, expected, actual, "actual output does not match expected output")
	})

	t.Run("composite literal", func(t *testing.T) {
		codeSample := `
		 	package main
		 	import "testing"
		 	func TestX(t *testing. T) {
				os.Args = []string{
					"fart",
				}
		 	}
		 `

		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[1].(*ast.FuncDecl).Body.List[0].(*ast.AssignStmt)

		actual := map[string]string{}
		expected := map[string]string{}

		parseAssignStmt(input, actual, map[string][]string{}, set.New())

		assert.Equal(t, expected, actual, "actual output does not match expected output")
	})
}

func TestParseHelperSelectorExpr(t *testing.T) {
	codeSample := `
		package main
		import "testing"

		func helperGenerator(t *testing.T) (ast.SelectorExpr, error) {
			return ast.SelectorExpr{}, nil
		}
	`

	p := parseChunkOfCode(t, codeSample)
	input := p.Decls[1].(*ast.FuncDecl).Type.Results.List[0].Type.(*ast.SelectorExpr)

	name := "arbitraryFunctionName"
	actual := map[string][]string{}
	expected := map[string][]string{
		name: {"ast.SelectorExpr"},
	}

	parseHelperSelectorExpr(input, name, actual)

	assert.Equal(t, expected, actual, "expected output did not match actual output")
}

func TestParseHelperFunction(t *testing.T) {
	t.Run("ident", func(t *testing.T) {
		codeSample := `
			package main
			import "testing"

			func helperGenerator(t *testing.T) (*Example, error) {
				return &Example{}, nil
			}
		`

		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[1].(*ast.FuncDecl)

		actual := map[string][]string{}
		expected := map[string][]string{
			"helperGenerator": {
				"Example",
				"error",
			},
		}

		parseHelperFunction(input, actual, set.New())

		assert.Equal(t, expected, actual, "expected output did not match actual output")
	})

	t.Run("selector", func(t *testing.T) {
		codeSample := `
			package main
			import "testing"

			func helperGenerator(t *testing.T) (ast.SelectorExpr, error) {
				return ast.SelectorExpr{}, nil
			}
		`

		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[1].(*ast.FuncDecl)

		actual := map[string][]string{}
		expected := map[string][]string{
			"helperGenerator": {
				"ast.SelectorExpr",
				"error",
			},
		}

		parseHelperFunction(input, actual, set.New())

		assert.Equal(t, expected, actual, "expected output did not match actual output")
	})

	t.Run("star selector", func(t *testing.T) {
		codeSample := `
			package main
			import "testing"

			func helperGenerator(t *testing.T) (*ast.SelectorExpr, error) {
				return &ast.SelectorExpr{}, nil
			}
		`

		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[1].(*ast.FuncDecl)

		actual := map[string][]string{}
		expected := map[string][]string{
			"helperGenerator": {
				"ast.SelectorExpr",
				"error",
			},
		}

		parseHelperFunction(input, actual, set.New())

		assert.Equal(t, expected, actual, "expected output did not match actual output")
	})
}

func TestParseFuncLit(t *testing.T) {
	codeSample := `
	package main
	import "testing"
	func TestX(t *testing. T) {
		subtest := func(t *testing.T) {
			var err error
			doSomeThings()
			err = doSomeOtherThings()
		}
		t.Run("subtest", subtest)
	}
`

	p := parseChunkOfCode(t, codeSample)
	input := p.Decls[1].(*ast.FuncDecl).Body.List[0].(*ast.AssignStmt).Rhs[0].(*ast.FuncLit)

	expected := set.New("doSomeThings", "doSomeOtherThings")
	actual := set.New()

	parseFuncLit(input, map[string]string{}, map[string][]string{}, actual)

	assert.Equal(t, expected, actual, "actual output does not match expected output")
}

func TestParseReturnStmt(t *testing.T) {
	codeSample := `
			package main
			func main(){
				return functionCall()
			}
		`

	p := parseChunkOfCode(t, codeSample)
	input := p.Decls[0].(*ast.FuncDecl).Body.List[0].(*ast.ReturnStmt)

	actual := set.New()
	expected := set.New("functionCall")

	parseReturnStmt(input, map[string]string{}, map[string][]string{}, actual)

	assert.Equal(t, expected, actual, "expected function name to be added to output")
}

func TestParseSelectStmt(t *testing.T) {
	codeSample := `
			package main
			func main(){
			temp := make(chan int)
			go func() {
				temp <- 0
			}()

			for {
				select {
				case <-temp:
					functionCall()
					return
				}
			}
			}
		`

	p := parseChunkOfCode(t, codeSample)
	input := p.Decls[0].(*ast.FuncDecl).Body.List[2].(*ast.ForStmt).Body.List[0].(*ast.SelectStmt)

	actual := set.New()
	expected := set.New("functionCall")

	parseSelectStmt(input, map[string]string{}, map[string][]string{}, actual)

	assert.Equal(t, expected, actual, "expected function name to be added to output")
}

func TestParseSendStmt(t *testing.T) {
	codeSample := `
			package main
			func main(){
				thing <- First()
				thing <- func(){
					Second()
				}()
				thing <- x.Third()
			}
		`

	p := parseChunkOfCode(t, codeSample)
	input := p.Decls[0].(*ast.FuncDecl).Body.List

	actual := set.New()
	expected := set.New(
		"First",
		"Second",
		"Example.Third",
	)

	for _, x := range input {
		in := x.(*ast.SendStmt)
		parseSendStmt(in, map[string]string{"x": "Example"}, map[string][]string{}, actual)
	}

	assert.Equal(t, expected, actual, "expected function name to be added to output")
}

func TestParseSwitchStmt(t *testing.T) {
	codeSample := `
	package main
	func main(){
		switch tmp {
		case tmp:
			functionCall()
		}
	}
`

	p := parseChunkOfCode(t, codeSample)
	input := p.Decls[0].(*ast.FuncDecl).Body.List[0].(*ast.SwitchStmt)

	actual := set.New()
	expected := set.New("functionCall")

	parseSwitchStmt(input, map[string]string{}, map[string][]string{}, actual)

	assert.Equal(t, expected, actual, "expected function name to be added to output")
}

func TestParseTypeSwitchStmt(t *testing.T) {
	codeSample := `
 			package main
 			func main(){
				func(i interface{}) {
					switch i.(type) {
					case string:
						functionCall()
					}
				}(tmp)
 			}
 		`

	p := parseChunkOfCode(t, codeSample)
	input := p.Decls[0].(*ast.FuncDecl).Body.List[0].(*ast.ExprStmt).X.(*ast.CallExpr).Fun.(*ast.FuncLit).Body.List[0].(*ast.TypeSwitchStmt)

	actual := set.New()
	expected := set.New("functionCall")

	parseTypeSwitchStmt(input, map[string]string{}, map[string][]string{}, actual)

	assert.Equal(t, expected, actual, "expected function name to be added to output")
}

func TestParseStmt(t *testing.T) {
	codeSample := `
		package main

		import "testing"

		func TestX(t *testing.T) string {
			// AssignStmt:
			tmp := "AssignStmt"
			foo := Foo{}
			bar := bar.Bar{}
			baz := &baz.Baz{}
			var x Example

			// RangeStmt
			for range [1]struct{}{} {
				A()
			}

			// IfStmt
			if true {
				B()
			}

			// DeclStmt
			var declStmt ast.DeclStmt

			// ExprStmt
			C()

			// DeferStmt
			defer func() {
				D()
			}()
			defer E()
			defer x.MethodOne()

			// ForStmt
			for i := 0; i < 1; i++ {
				F()
			}

			// GoStmt
			go G()
			go func() {
				H()
			}()
			go x.MethodTwo()

			// SelectStmt
			temp := make(chan int)
			go func() {
				temp <- 0
			}()

			for {
				select {
				case <-temp:
					I()
					return
				}
			}

			// SendStmt
			thing <- J()
			thing <- func(){
				K()
			}()
			thing <- x.MethodThree()

			// Args
			assert.True(t, x.MethodFour())
			assert.True(t, x.MethodFive())

			// SwitchStmt
			switch tmp {
			case tmp:
				L()
			}

			// TypeSwitchStmt
			func(i interface{}) {
				switch i.(type) {
				case string:
					M()
				}
			}(tmp)

			// miscellany
			n := []string{
				N(),
			}

			o := &Example{
				o: O(),
			}

			p := &Example{
				P(),
			}

			// ReturnStmt
			return Q()
		}
	`

	actual := set.New("make")
	expected := set.New(
		"A",
		"B",
		"C",
		"D",
		"E",
		"F",
		"G",
		"H",
		"I",
		"J",
		"K",
		"L",
		"M",
		"N",
		"O",
		"P",
		"Q",
		"make",
		"Example.MethodOne",
		"Example.MethodTwo",
		"Example.MethodThree",
		"Example.MethodFour",
		"Example.MethodFive",
	)

	p := parseChunkOfCode(t, codeSample)
	for _, input := range p.Decls[1].(*ast.FuncDecl).Body.List {
		parseStmt(input, map[string]string{"x": "Example"}, map[string][]string{}, actual)
	}

	diff := set.StringSlice(set.Difference(expected, actual))
	assert.Empty(t, diff, "diff should be empty")
	assert.Equal(t, expected, actual, "actual output does not match expected output")
}

func TestGetDeclaredNames(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		in, err := parser.ParseFile(token.NewFileSet(), "example_packages/simple/main.go", nil, parser.AllErrors)
		if err != nil {
			t.Logf("failing because ParseFile returned error: %v", err)
			t.FailNow()
		}

		expected := map[string]blanketFunc{
			"a":       {Name: "a"},
			"b":       {Name: "b"},
			"c":       {Name: "c"},
			"wrapper": {Name: "wrapper"},
		}
		actual := map[string]blanketFunc{}

		getDeclaredNames(in, token.NewFileSet(), actual)

		assert.Equal(t, expected, actual, "expected output did not match actual output")
	})

	t.Run("methods", func(t *testing.T) {
		in, err := parser.ParseFile(token.NewFileSet(), "example_packages/methods/main.go", nil, parser.AllErrors)
		if err != nil {
			t.Logf("failing because ParseFile returned error: %v", err)
			t.FailNow()
		}

		expected := map[string]blanketFunc{
			"example.A": {Name: "example.A"},
			"example.B": {Name: "example.B"},
			"example.C": {Name: "example.C"},
			"example.D": {Name: "example.D"},
			"example.E": {Name: "example.E"},
			"example.F": {Name: "example.F"},
			"wrapper":   {Name: "wrapper"},
		}
		actual := map[string]blanketFunc{}

		getDeclaredNames(in, token.NewFileSet(), actual)

		assert.Equal(t, expected, actual, "expected output did not match actual output")
	})
}

func TestGetCalledNames(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		in, err := parser.ParseFile(token.NewFileSet(), "example_packages/simple/main_test.go", nil, parser.AllErrors)
		if err != nil {
			t.Logf("failing because ParseFile returned error: %v", err)
			t.FailNow()
		}

		expectedDeclarations := []string{"a", "c", "wrapper"}
		expected := set.New()
		for _, x := range expectedDeclarations {
			expected.Add(x)
		}

		actual := set.New()

		getCalledNames(in, map[string]string{}, map[string][]string{}, actual)

		assert.Equal(t, expected, actual, "expected output did not match actual output")
	})

	t.Run("methods", func(t *testing.T) {
		in, err := parser.ParseFile(token.NewFileSet(), "example_packages/methods/main_test.go", nil, parser.AllErrors)
		if err != nil {
			t.Logf("failing because ParseFile returned error: %v", err)
			t.FailNow()
		}

		expected := set.New(
			"example.A",
			"example.B",
			"example.C",
			"example.D",
			"example.E",
			"wrapper",
		)
		actual := set.New()

		helperFunctionMap := map[string][]string{
			"helperGenerator": {
				"example",
				"error",
			},
		}
		getCalledNames(in, map[string]string{}, helperFunctionMap, actual)

		assert.Equal(t, expected, actual, "expected output did not match actual output")
	})
}

func TestFindHelperFuncs(t *testing.T) {
	in, err := parser.ParseFile(token.NewFileSet(), "example_packages/methods/main_test.go", nil, parser.AllErrors)
	if err != nil {
		t.Logf("failing because ParseFile returned error: %v", err)
		t.FailNow()
	}

	expected := map[string][]string{
		"helperGenerator": {
			"example",
			"error",
		},
	}
	actual := map[string][]string{}
	findHelperFuncs(in, actual, set.New())

	assert.Equal(t, expected, actual, "expected output did not match actual output")
}

func TestAnalyze(t *testing.T) {
	debug = true
	simpleMainPath := fmt.Sprintf("%s/main.go", buildExamplePackagePath(t, "simple", true))
	expected := blanketReport{
		DeclaredDetails: map[string]blanketFunc{
			"a": {
				Name:     "a",
				Filename: simpleMainPath,
				DeclPos: token.Position{
					Filename: simpleMainPath,
					Offset:   16,
					Line:     3,
					Column:   1,
				},
				RBracePos: token.Position{
					Filename: simpleMainPath,
					Offset:   32,
					Line:     3,
					Column:   17,
				},
				LBracePos: token.Position{
					Filename: simpleMainPath,
					Offset:   46,
					Line:     5,
					Column:   1,
				},
			},
			"b": {
				Name:     "b",
				Filename: simpleMainPath,
				DeclPos: token.Position{
					Filename: simpleMainPath,
					Offset:   49,
					Line:     7,
					Column:   1,
				},
				RBracePos: token.Position{
					Filename: simpleMainPath,
					Offset:   65,
					Line:     7,
					Column:   17,
				},
				LBracePos: token.Position{
					Filename: simpleMainPath,
					Offset:   79,
					Line:     9,
					Column:   1,
				},
			},
			"c": {
				Name:     "c",
				Filename: simpleMainPath,
				DeclPos: token.Position{
					Filename: simpleMainPath,
					Offset:   82,
					Line:     11,
					Column:   1,
				},
				RBracePos: token.Position{
					Filename: simpleMainPath,
					Offset:   98,
					Line:     11,
					Column:   17,
				},
				LBracePos: token.Position{
					Filename: simpleMainPath,
					Offset:   112,
					Line:     13,
					Column:   1,
				},
			},
			"wrapper": {
				Name:     "wrapper",
				Filename: simpleMainPath,
				DeclPos: token.Position{
					Filename: simpleMainPath,
					Offset:   115,
					Line:     15,
					Column:   1,
				},
				RBracePos: token.Position{
					Filename: simpleMainPath,
					Offset:   130,
					Line:     15,
					Column:   16,
				},
				LBracePos: token.Position{
					Filename: simpleMainPath,
					Offset:   147,
					Line:     19,
					Column:   1,
				},
			},
		},
		Called:   set.New("a", "c", "wrapper"),
		Declared: set.New("a", "b", "c", "wrapper"),
	}
	examplePath := buildExamplePackagePath(t, "simple", false)
	actual := analyze(examplePath)

	assert.Equal(t, expected, actual, "expected output did not match actual output")

}
