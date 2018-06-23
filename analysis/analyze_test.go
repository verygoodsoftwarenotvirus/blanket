package analysis

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"testing"

	"github.com/verygoodsoftwarenotvirus/blanket/lib/util"

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

	t.Run("ident", func(_t *testing.T) {
		analyzer := NewAnalyzer()

		codeSample := `
			package main

			func main() {
				functionCall()
			}
		`

		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[0].(*ast.FuncDecl).Body.List[0].(*ast.ExprStmt).X.(*ast.CallExpr).Fun

		expected := set.New("init", "functionCall")

		analyzer.parseExpr(input)

		assert.Equal(t, expected, analyzer.calledFuncs, "expected function name to be added to output")
	})

	t.Run("selector", func(_t *testing.T) {
		analyzer := NewAnalyzer()
		analyzer.nameToTypeMap["class"] = "Example"

		codeSample := `
			package main

			func main() {
				class.methodCall()
			}
		`

		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[0].(*ast.FuncDecl).Body.List[0].(*ast.ExprStmt).X.(*ast.CallExpr).Fun
		expected := set.New("init", "Example.methodCall")

		analyzer.parseExpr(input)

		assert.Equal(t, expected, analyzer.calledFuncs, "expected function name to be added to output")
	})

	t.Run("function literal", func(_t *testing.T) {
		analyzer := NewAnalyzer()

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
		expected := set.New("init", "functionCall")

		analyzer.parseExpr(input)

		assert.Equal(t, expected, analyzer.calledFuncs, "expected function name to be added to output")
	})
}

func TestParseCallExpr(t *testing.T) {
	t.Run("with ast.Ident", func(_t *testing.T) {
		analyzer := NewAnalyzer()

		codeSample := `
			package main
			var function func()
			func main(){
				fart := function()
			}
		`

		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[1].(*ast.FuncDecl).Body.List[0].(*ast.AssignStmt).Rhs[0].(*ast.CallExpr)
		expected := set.New("init", "function")

		analyzer.parseCallExpr(input)

		assert.Equal(t, expected, analyzer.calledFuncs, "expected function name to be added to output")
	})

	t.Run("with ast.SelectorExpr", func(_t *testing.T) {
		analyzer := NewAnalyzer()
		analyzer.nameToTypeMap["s"] = "Struct"

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
		expected := set.New("init", "Struct.method")

		analyzer.parseCallExpr(input)

		assert.Equal(t, expected, analyzer.calledFuncs, "expected function name to be added to output")
	})

	t.Run("with ast.SelectorExpr, but no matching entity", func(_t *testing.T) {
		analyzer := NewAnalyzer()

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
		expected := set.New("init")

		analyzer.parseCallExpr(input)

		assert.Equal(t, expected, analyzer.calledFuncs, "expected function name to be added to output")
	})

	t.Run("with funcLit in argument list", func(_t *testing.T) {
		analyzer := NewAnalyzer()

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
		expected := set.New("init", "arbitraryCallExpression")

		analyzer.parseCallExpr(input)

		assert.Equal(t, expected, analyzer.calledFuncs, "expected function name to be added to output")
	})
}

func TestParseUnaryExpr(t *testing.T) {
	analyzer := NewAnalyzer()

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

	analyzer.parseUnaryExpr(input, "s")

	assert.Equal(t, expected, analyzer.nameToTypeMap, "actual output does not match expected output")
}

func TestParseDeclStmt(t *testing.T) {
	analyzer := NewAnalyzer()

	codeSample := `
		package main
		func main(){
			var test bool
		}
	`

	p := parseChunkOfCode(t, codeSample)
	input := p.Decls[0].(*ast.FuncDecl).Body.List[0].(*ast.DeclStmt)
	expected := map[string]string{"test": "bool"}

	analyzer.parseDeclStmt(input)

	assert.Equal(t, expected, analyzer.nameToTypeMap, "actual output does not match expected output")
}

func TestParseExprStmt(t *testing.T) {
	t.Run("CallExpr.Fun.(*ast.Ident)", func(_t *testing.T) {
		analyzer := NewAnalyzer()

		codeSample := `
			package main
			var example func()
			func main(){
				example()
			}
		`
		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[1].(*ast.FuncDecl).Body.List[0].(*ast.ExprStmt)
		expected := set.New("init", "example")

		analyzer.parseExprStmt(input)

		assert.Equal(t, expected, analyzer.calledFuncs, "actual output does not match expected output")
	})

	t.Run("CallExpr.Fun.(*ast.Selector)", func(_t *testing.T) {
		analyzer := NewAnalyzer()
		analyzer.nameToTypeMap["e"] = "Example"

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
		expected := set.New("init", "Example.method")

		analyzer.parseExprStmt(input)

		assert.Equal(t, expected, analyzer.calledFuncs, "actual output does not match expected output")
	})
}

func TestParseCompositeLit(t *testing.T) {
	t.Run("ident", func(_t *testing.T) {
		analyzer := NewAnalyzer()

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
		expected := set.New("init", "methodCallAsArg")

		analyzer.parseCompositeLit(input, "e")

		assert.Equal(t, expected, analyzer.calledFuncs, "actual output does not match expected output")
	})

	t.Run("selector", func(_t *testing.T) {
		analyzer := NewAnalyzer()
		analyzer.nameToTypeMap["e"] = "Example"

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
		expected := set.New("init", "Example.methodCallAsArg")

		analyzer.parseCompositeLit(input, "e")

		assert.Equal(t, expected, analyzer.calledFuncs, "actual output does not match expected output")
	})
}

func TestParseGenDecl(t *testing.T) {
	analyzer := NewAnalyzer()

	codeSample := `
		package main
		var thing string
		func main(){}
	`

	p := parseChunkOfCode(t, codeSample)
	input := p.Decls[0].(*ast.GenDecl)
	expected := map[string]string{"thing": "string"}

	analyzer.parseGenDecl(input)

	assert.Equal(t, expected, analyzer.nameToTypeMap, "expected function name to be added to output")
}

func TestParseFuncDecl(t *testing.T) {
	t.Run("simple", func(_t *testing.T) {
		analyzer := NewAnalyzer()

		codeSample := `
			package test
			func example(){}
		`

		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[0].(*ast.FuncDecl)

		expected := "example"
		actual := analyzer.parseFuncDecl(input)

		assert.Equal(t, expected, actual, "actual output does not match expected output")
	})

	t.Run("with receiver", func(_t *testing.T) {
		analyzer := NewAnalyzer()

		codeSample := `
			package test
			type Example struct{}
			func (e Example) method(){}
		`

		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[1].(*ast.FuncDecl)

		expected := "Example.method"
		actual := analyzer.parseFuncDecl(input)

		assert.Equal(t, expected, actual, "actual output does not match expected output")
	})

	t.Run("with ptr receiver", func(_t *testing.T) {
		analyzer := NewAnalyzer()

		codeSample := `
			package test
			type Example struct{}
			func (e *Example) method(){}
		`

		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[1].(*ast.FuncDecl)

		expected := "Example.method"
		actual := analyzer.parseFuncDecl(input)

		assert.Equal(t, expected, actual, "actual output does not match expected output")
	})
}

func TestParseAssignStmt(t *testing.T) {
	t.Run("CallExpr", func(_t *testing.T) {
		analyzer := NewAnalyzer()

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
		expected := set.New("init", "example")

		analyzer.parseAssignStmt(input)

		assert.Equal(t, expected, analyzer.calledFuncs, "actual output does not match expected output")
	})

	t.Run("CallExpr with multiple returns and ast.Ident Fun value", func(_t *testing.T) {
		analyzer := NewAnalyzer()
		analyzer.helperFunctionReturnMap["example"] = []string{"X", "Y"}

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
		expected := map[string]string{
			"x": "X",
			"y": "Y",
		}

		analyzer.parseAssignStmt(input)

		assert.Equal(t, expected, analyzer.nameToTypeMap, "actual output does not match expected output")
	})

	t.Run("Assign statement with multiple returns from some external function", func(_t *testing.T) {
		analyzer := NewAnalyzer()

		codeSample := `
			package main
			import "testing"
			func TestX() {
				req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
			}
		`

		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[1].(*ast.FuncDecl).Body.List[0].(*ast.AssignStmt)
		expected := map[string]string{}

		analyzer.parseAssignStmt(input)

		assert.Equal(t, expected, analyzer.nameToTypeMap, "actual output does not match expected output")
	})

	t.Run("Assign statement with multiple returns from some internal function", func(_t *testing.T) {
		analyzer := NewAnalyzer()
		analyzer.helperFunctionReturnMap["someHelperFunctionForTestsOnly"] = []string{"http.Request", "error"}

		codeSample := `
			package main
			import "testing"
			func TestX() {
				req, err := someHelperFunctionForTestsOnly()
			}
		`

		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[1].(*ast.FuncDecl).Body.List[0].(*ast.AssignStmt)
		expected := map[string]string{
			"req": "http.Request",
			"err": "error",
		}

		analyzer.parseAssignStmt(input)

		assert.Equal(t, expected, analyzer.nameToTypeMap, "actual output does not match expected output")
	})

	t.Run("UnaryExpr", func(_t *testing.T) {
		analyzer := NewAnalyzer()

		codeSample := `
			package main
			import "testing"
			func TestX(t *testing.T) {
				test := &SomeStruct{}
			}
		`

		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[1].(*ast.FuncDecl).Body.List[0].(*ast.AssignStmt)

		expected := map[string]string{
			"test": "SomeStruct",
		}

		analyzer.parseAssignStmt(input)

		assert.Equal(t, expected, analyzer.nameToTypeMap, "actual output does not match expected output")
	})

	t.Run("multiple unary expressions", func(_t *testing.T) {
		analyzer := NewAnalyzer()

		codeSample := `
			package main
			import "testing"
			func TestX(t *testing.T) {
				one, other := &SomeStruct{}, &SomeOtherStruct{}
			}
		`

		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[1].(*ast.FuncDecl).Body.List[0].(*ast.AssignStmt)
		expected := map[string]string{
			"one":   "SomeStruct",
			"other": "SomeOtherStruct",
		}

		analyzer.parseAssignStmt(input)

		assert.Equal(t, expected, analyzer.nameToTypeMap, "actual output does not match expected output")
	})

	t.Run("FuncLit", func(_t *testing.T) {
		analyzer := NewAnalyzer()

		codeSample := `
		 	package main
		 	import "testing"
		 	func TestX(t *testing. T) {
		 		subtest := func(_t *testing.T) {
	analyzer := NewAnalyzer()
}
		 		t.Run("subtest", subtest)
		 	}
		 `

		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[1].(*ast.FuncDecl).Body.List[0].(*ast.AssignStmt)
		expected := map[string]string{}

		analyzer.parseAssignStmt(input)

		assert.Equal(t, expected, analyzer.nameToTypeMap, "actual output does not match expected output")
	})

	t.Run("composite literal", func(_t *testing.T) {
		analyzer := NewAnalyzer()

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
		expected := map[string]string{}

		analyzer.parseAssignStmt(input)

		assert.Equal(t, expected, analyzer.nameToTypeMap, "actual output does not match expected output")
	})
}

func TestParseHelperSelectorExpr(t *testing.T) {
	analyzer := NewAnalyzer()

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
	expected := map[string][]string{
		name: {"ast.SelectorExpr"},
	}

	analyzer.parseHelperSelectorExpr(input, name)

	assert.Equal(t, expected, analyzer.helperFunctionReturnMap, "expected output did not match actual output")
}

func TestParseHelperFunction(t *testing.T) {
	t.Run("ident", func(_t *testing.T) {
		analyzer := NewAnalyzer()

		codeSample := `
			package main
			import "testing"

			func helperGenerator(t *testing.T) (*Example, error) {
				return &Example{}, nil
			}
		`

		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[1].(*ast.FuncDecl)

		expected := map[string][]string{
			"helperGenerator": {
				"Example",
				"error",
			},
		}
		analyzer.parseHelperFunction(input)

		assert.Equal(t, expected, analyzer.helperFunctionReturnMap, "expected output did not match actual output")
	})

	t.Run("selector", func(_t *testing.T) {
		analyzer := NewAnalyzer()

		codeSample := `
			package main
			import "testing"

			func helperGenerator(t *testing.T) (ast.SelectorExpr, error) {
				return ast.SelectorExpr{}, nil
			}
		`

		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[1].(*ast.FuncDecl)

		expected := map[string][]string{
			"helperGenerator": {
				"ast.SelectorExpr",
				"error",
			},
		}

		analyzer.parseHelperFunction(input)

		assert.Equal(t, expected, analyzer.helperFunctionReturnMap, "expected output did not match actual output")
	})

	t.Run("star selector", func(_t *testing.T) {
		analyzer := NewAnalyzer()

		codeSample := `
			package main
			import "testing"

			func helperGenerator(t *testing.T) (*ast.SelectorExpr, error) {
				return &ast.SelectorExpr{}, nil
			}
		`

		p := parseChunkOfCode(t, codeSample)
		input := p.Decls[1].(*ast.FuncDecl)

		expected := map[string][]string{
			"helperGenerator": {
				"ast.SelectorExpr",
				"error",
			},
		}

		analyzer.parseHelperFunction(input)

		assert.Equal(t, expected, analyzer.helperFunctionReturnMap, "expected output did not match actual output")
	})
}

func TestParseFuncLit(t *testing.T) {
	analyzer := NewAnalyzer()

	codeSample := `
	package main
	import "testing"
	func TestX(t *testing. T) {
		subtest := func(_t *testing.T) {
			var err error
			doSomeThings()
			err = doSomeOtherThings()
		}
		t.Run("subtest", subtest)
	}
`

	p := parseChunkOfCode(t, codeSample)
	input := p.Decls[1].(*ast.FuncDecl).Body.List[0].(*ast.AssignStmt).Rhs[0].(*ast.FuncLit)
	expected := set.New("init", "doSomeThings", "doSomeOtherThings")

	analyzer.parseFuncLit(input)

	assert.Equal(t, expected, analyzer.calledFuncs, "actual output does not match expected output")
}

func TestParseReturnStmt(t *testing.T) {
	analyzer := NewAnalyzer()

	codeSample := `
			package main
			func main(){
				return functionCall()
			}
		`

	p := parseChunkOfCode(t, codeSample)
	input := p.Decls[0].(*ast.FuncDecl).Body.List[0].(*ast.ReturnStmt)

	expected := set.New("init", "functionCall")

	analyzer.parseReturnStmt(input)

	assert.Equal(t, expected, analyzer.calledFuncs, "expected function name to be added to output")
}

func TestParseSelectStmt(t *testing.T) {
	analyzer := NewAnalyzer()

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
	expected := set.New("init", "functionCall")

	analyzer.parseSelectStmt(input)

	assert.Equal(t, expected, analyzer.calledFuncs, "expected function name to be added to output")
}

func TestParseSendStmt(t *testing.T) {
	analyzer := NewAnalyzer()
	analyzer.nameToTypeMap["x"] = "Example"

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
	expected := set.New(
		"init",
		"First",
		"Second",
		"Example.Third",
	)

	for _, x := range input {
		in := x.(*ast.SendStmt)
		analyzer.parseSendStmt(in)
	}

	assert.Equal(t, expected, analyzer.calledFuncs, "expected function name to be added to output")
}

func TestParseSwitchStmt(t *testing.T) {
	analyzer := NewAnalyzer()

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
	expected := set.New("init", "functionCall")

	analyzer.parseSwitchStmt(input)

	assert.Equal(t, expected, analyzer.calledFuncs, "expected function name to be added to output")
}

func TestParseTypeSwitchStmt(t *testing.T) {
	analyzer := NewAnalyzer()

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
	expected := set.New("init", "functionCall")

	analyzer.parseTypeSwitchStmt(input)

	assert.Equal(t, expected, analyzer.calledFuncs, "expected function name to be added to output")
}

func TestParseStmt(t *testing.T) {
	analyzer := NewAnalyzer()

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

	//analyzer.calledFuncs.Add("make")
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
		"init",
		"make",
		"Example.MethodOne",
		"Example.MethodTwo",
		"Example.MethodThree",
		"Example.MethodFour",
		"Example.MethodFive",
	)

	p := parseChunkOfCode(t, codeSample)
	for _, input := range p.Decls[1].(*ast.FuncDecl).Body.List {
		analyzer.parseStmt(input)
	}

	diff := set.StringSlice(set.Difference(expected, analyzer.calledFuncs))
	assert.Empty(t, diff, "diff should be empty")
	assert.Equal(t, expected, analyzer.calledFuncs, "actual output does not match expected output")
}

func TestGetDeclaredNames(t *testing.T) {
	t.Run("simple", func(_t *testing.T) {
		analyzer := NewAnalyzer()

		in, err := parser.ParseFile(token.NewFileSet(), "../example_packages/simple/main.go", nil, parser.AllErrors)
		if err != nil {
			t.Logf("failing because ParseFile returned error: %v", err)
			t.FailNow()
		}

		expected := map[string]BlanketFunc{
			"a":       {Name: "a"},
			"b":       {Name: "b"},
			"c":       {Name: "c"},
			"wrapper": {Name: "wrapper"},
		}

		analyzer.getDeclaredNames(in)

		assert.Equal(t, expected, analyzer.declaredFuncInfo, "expected output did not match actual output")
	})

	t.Run("methods", func(_t *testing.T) {
		analyzer := NewAnalyzer()

		in, err := parser.ParseFile(token.NewFileSet(), "../example_packages/methods/main.go", nil, parser.AllErrors)
		if err != nil {
			t.Logf("failing because ParseFile returned error: %v", err)
			t.FailNow()
		}

		expected := map[string]BlanketFunc{
			"example.A": {Name: "example.A"},
			"example.B": {Name: "example.B"},
			"example.C": {Name: "example.C"},
			"example.D": {Name: "example.D"},
			"example.E": {Name: "example.E"},
			"example.F": {Name: "example.F"},
			"wrapper":   {Name: "wrapper"},
		}
		analyzer.getDeclaredNames(in)

		assert.Equal(t, expected, analyzer.declaredFuncInfo, "expected output did not match actual output")
	})
}

func TestGetCalledNames(t *testing.T) {
	t.Run("simple", func(_t *testing.T) {
		analyzer := NewAnalyzer()

		in, err := parser.ParseFile(token.NewFileSet(), "../example_packages/simple/main_test.go", nil, parser.AllErrors)
		if err != nil {
			t.Logf("failing because ParseFile returned error: %v", err)
			t.FailNow()
		}

		expectedDeclarations := []string{"a", "c", "wrapper", "init"}
		expected := set.New()
		for _, x := range expectedDeclarations {
			expected.Add(x)
		}

		analyzer.getCalledNames(in)

		assert.Equal(t, expected, analyzer.calledFuncs, "expected output did not match actual output")
	})

	t.Run("methods", func(_t *testing.T) {
		analyzer := NewAnalyzer()
		analyzer.helperFunctionReturnMap["helperGenerator"] = []string{"example", "error"}

		in, err := parser.ParseFile(token.NewFileSet(), "../example_packages/methods/main_test.go", nil, parser.AllErrors)
		if err != nil {
			t.Logf("failing because ParseFile returned error: %v", err)
			t.FailNow()
		}

		expected := set.New(
			"example.A",
			"example.B",
			"example.C",
			"helperGenerator",
			"example.D",
			"example.E",
			"wrapper",
			"init",
		)
		analyzer.getCalledNames(in)

		use := func(...interface{}) {}
		x := analyzer.calledFuncs.List()
		use(x)

		assert.Equal(t, expected, analyzer.calledFuncs, "expected output did not match actual output")
	})
}

func TestFindHelperFuncs(t *testing.T) {
	analyzer := NewAnalyzer()

	in, err := parser.ParseFile(token.NewFileSet(), "../example_packages/methods/main_test.go", nil, parser.AllErrors)
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
	analyzer.findHelperFuncs(in)

	assert.Equal(t, expected, analyzer.helperFunctionReturnMap, "expected output did not match actual output")
}

func TestAnalyze(t *testing.T) {
	analyzer := NewAnalyzer()
	analyzer.debug = true

	simpleMainPath := fmt.Sprintf("%s/main.go", util.BuildExamplePackagePath(t, "simple", true))
	expected := &BlanketReport{
		DeclaredDetails: map[string]BlanketFunc{
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
	examplePath := util.BuildExamplePackagePath(t, "simple", false)
	actual, err := analyzer.Analyze(examplePath)

	assert.NoError(t, err, "Analyze produced an unexpected error")
	assert.Equal(t, expected, actual, "expected output did not match actual output")
}

func TestGenerateDiffReport(t *testing.T) {
	analyzer := NewAnalyzer()
	simpleMainPath := fmt.Sprintf("%s/main.go", util.BuildExamplePackagePath(t, "simple", true))
	exampleReport := &BlanketReport{
		DeclaredDetails: map[string]BlanketFunc{
			"A": {
				Name:     "A",
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
			"B": {
				Name:     "B",
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
			"C": {
				Name:     "C",
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
		Called:   set.New("A", "C", "wrapper"),
		Declared: set.New("A", "B", "C", "wrapper"),
	}

	expected := &blanketOutput{
		LongestFunctionNameLength: 1,
		DeclaredCount:             4,
		CalledCount:               3,
		Score:                     75,
		Details: map[string][]BlanketFunc{
			simpleMainPath: {
				BlanketFunc{
					Name:     "B",
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
			},
		},
	}
	analyzer.latestReport = exampleReport
	analyzer.declaredFuncInfo = exampleReport.DeclaredDetails
	actual := analyzer.GenerateDiffReport()

	print()

	assert.Equal(t, expected, actual, "expected and actual diff reports should match.")
}
