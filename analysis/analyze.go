package analysis

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/fatih/set"
	"github.com/pkg/errors"
)

type analyzer struct {
	fileset                 *token.FileSet
	debug                   bool
	declaredFuncInfo        map[string]BlanketFunc
	calledFuncs             *set.Set
	helperFunctionReturnMap map[string][]string
	nameToTypeMap           map[string]string
	latestReport            *BlanketReport
}

func (a *analyzer) parseExpr(in ast.Expr) {
	// FIXME: iterate over in.Args to see if there are function calls
	switch f := in.(type) {
	case *ast.Ident:
		functionName := f.Name
		if _, ok := a.helperFunctionReturnMap[functionName]; !ok {
			a.calledFuncs.Add(functionName)
		}
	case *ast.SelectorExpr:
		if x, ok := f.X.(*ast.Ident); ok {
			structVarName := x.Name
			calledMethodName := f.Sel.Name
			if _, ok := a.nameToTypeMap[structVarName]; ok {
				a.calledFuncs.Add(fmt.Sprintf("%s.%s", a.nameToTypeMap[structVarName], calledMethodName))
			}
		}
	case *ast.FuncLit:
		a.parseFuncLit(f)
	}
}

func (a *analyzer) parseCallExpr(in *ast.CallExpr) {
	for _, arg := range in.Args {
		switch at := arg.(type) {
		// case *ast.Ident:
		case *ast.CallExpr:
			a.parseCallExpr(at)
		case *ast.FuncLit:
			a.parseFuncLit(at)
		}
	}
	a.parseExpr(in.Fun)
}

// parseUnaryExpr parses Unary expressions. From the go/ast docs:
//      A UnaryExpr node represents a unary expression. Unary "*" expressions are represented via StarExpr nodes.
// (handles declarations like `callExpr := &ast.UnaryExpr{}` or `callExpr := ast.UnaryExpr{}`)
func (a *analyzer) parseUnaryExpr(in *ast.UnaryExpr, varName string) {
	if cl, ok := in.X.(*ast.CompositeLit); ok {
		for _, e := range cl.Elts {
			switch et := e.(type) {
			case *ast.CallExpr:
				a.parseExpr(et.Fun)
			case *ast.KeyValueExpr:
				if vt, ok := et.Value.(*ast.CallExpr); ok {
					a.parseCallExpr(vt)
				}
			}
		}
		switch u := cl.Type.(type) {
		case *ast.Ident:
			a.nameToTypeMap[varName] = u.Name
		case *ast.SelectorExpr:
			a.nameToTypeMap[varName] = u.Sel.Name
		}
	}
}

// parseDeclStmt parses declaration statments. From the go/ast docs:
// 		A DeclStmt node represents a declaration in a statement list.
// DeclStmts come from function bodies, GenDecls come from package-wide const or var declarations
func (a *analyzer) parseDeclStmt(in *ast.DeclStmt) {
	if gd, ok := in.Decl.(*ast.GenDecl); ok {
		if len(gd.Specs) > 0 {
			if s, ok := gd.Specs[0].(*ast.ValueSpec); ok {
				if len(s.Names) > 0 {
					varName := s.Names[0].Name
					switch t := s.Type.(type) {
					case *ast.Ident:
						a.nameToTypeMap[varName] = t.Name
					case *ast.SelectorExpr:
						a.nameToTypeMap[varName] = t.Sel.Name
					}
				}
			}
		}
	}
}

// parseExprStmt parses expression statements. From the go/ast docs:
// 		An ExprStmt node represents a (stand-alone) expression in a statement list.
func (a *analyzer) parseExprStmt(in *ast.ExprStmt) {
	if c, ok := in.X.(*ast.CallExpr); ok {
		a.parseCallExpr(c)
	}
}

func (a *analyzer) parseCompositeLit(in *ast.CompositeLit, varName string) {
	for _, e := range in.Elts {
		if et, ok := e.(*ast.CallExpr); ok {
			a.parseExpr(et.Fun)
		}
	}

	switch t := in.Type.(type) {
	case *ast.Ident:
		a.nameToTypeMap[varName] = t.Name
	case *ast.SelectorExpr:
		a.nameToTypeMap[varName] = t.Sel.Name
	}
}

// parseGenDecl handles GenDecl nodes. From the go/ast docs:
//     A GenDecl node (generic declaration node) represents an import, constant, type or variable declaration.
func (a *analyzer) parseGenDecl(in *ast.GenDecl) {
	for _, spec := range in.Specs {
		if global, ok := spec.(*ast.ValueSpec); ok {
			if len(global.Names) > 0 {
				varName := global.Names[0].Name
				if global.Type != nil {
					if t, ok := global.Type.(*ast.Ident); ok {
						typeName := t.Name
						a.nameToTypeMap[varName] = typeName
					}
				}
			}
		}
	}
}

// parseFuncDecl parses function declarations. From the go/ast docs:
//		A FuncDecl node represents a function declaration.
// the main purpose of this is to parse functions that are declared in non-test go files.
func (a *analyzer) parseFuncDecl(f *ast.FuncDecl) string {
	functionName := f.Name.Name // "Avoid Stutter" lol
	var parentName string
	if f.Recv != nil {
		if len(f.Recv.List) > 0 {
			switch x := f.Recv.List[0].Type.(type) {
			case *ast.StarExpr:
				if parent, ok := x.X.(*ast.Ident); ok {
					parentName = parent.Name
				}
			case *ast.Ident:
				parentName = x.Obj.Name
			}
		}
	}

	if parentName != "" {
		return fmt.Sprintf("%s.%s", parentName, functionName)
	}
	return functionName
}

// parseAssignStmt handles AssignStmt nodes. From the go/ast docs:
//    An AssignStmt node represents an assignment or a short variable declaration
func (a *analyzer) parseAssignStmt(in *ast.AssignStmt) {
	leftHandSide := []string{}
	for i := range in.Lhs {
		if l, ok := in.Lhs[i].(*ast.Ident); ok {
			varName := l.Name
			leftHandSide = append(leftHandSide, varName)
		}
	}

	for j := range in.Rhs {
		switch t := in.Rhs[j].(type) {
		case *ast.FuncLit:
			a.parseFuncLit(t)
		case *ast.UnaryExpr:
			a.parseUnaryExpr(t, leftHandSide[j])
		case *ast.CompositeLit:
			if len(leftHandSide) > j {
				a.parseCompositeLit(t, leftHandSide[j])
			} else {
				a.parseCompositeLit(t, "")
			}
		case *ast.CallExpr:
			var functionName string
			switch funcInfo := t.Fun.(type) {
			case *ast.Ident:
				functionName = funcInfo.Name
			case *ast.SelectorExpr:
				functionName = funcInfo.Sel.Name
			}
			if _, ok := a.helperFunctionReturnMap[functionName]; ok {
				for i, thing := range leftHandSide {
					a.nameToTypeMap[thing] = a.helperFunctionReturnMap[functionName][i]
				}
			}

			a.parseCallExpr(t)
		}
	}
}

func (a *analyzer) parseHelperSelectorExpr(in *ast.SelectorExpr, functionName string) {
	if pkg, ok := in.X.(*ast.Ident); ok {
		pkgName := pkg.Name
		pkgStruct := in.Sel.Name
		a.helperFunctionReturnMap[functionName] = append(a.helperFunctionReturnMap[functionName], fmt.Sprintf("%s.%s", pkgName, pkgStruct))
	}
}

func (a *analyzer) parseHelperFunction(in *ast.FuncDecl) {
	functionName := in.Name.Name
	if in.Type.Results != nil {
		for _, r := range in.Type.Results.List {
			switch rt := r.Type.(type) {
			case *ast.SelectorExpr:
				a.parseHelperSelectorExpr(rt, functionName)
			case *ast.StarExpr:
				switch x := rt.X.(type) {
				case *ast.Ident:
					a.helperFunctionReturnMap[functionName] = append(a.helperFunctionReturnMap[functionName], x.Name)
				case *ast.SelectorExpr:
					a.parseHelperSelectorExpr(x, functionName)
				}
			case *ast.Ident:
				a.helperFunctionReturnMap[functionName] = append(a.helperFunctionReturnMap[functionName], rt.Name)
			}
		}
	}
}

// parseFuncLit parses a function literal. From the go/ast docs:
// 		A FuncLit node represents a function literal.
// FuncLits have bodies that we basically need to explore the same way that we explore a normal function.
func (a *analyzer) parseFuncLit(in *ast.FuncLit) {
	for _, le := range in.Body.List {
		a.parseStmt(le)
	}
}

func (a *analyzer) parseReturnStmt(in *ast.ReturnStmt) {
	for _, x := range in.Results {
		if y, ok := x.(*ast.CallExpr); ok {
			a.parseExpr(y.Fun)
		}
	}
}

func (a *analyzer) parseSelectStmt(in *ast.SelectStmt) {
	for _, x := range in.Body.List {
		if y, ok := x.(*ast.CommClause); ok {
			for _, z := range y.Body {
				a.parseStmt(z)
			}
		}
	}
}

// parseSendStmt parses a send statement. (<-)
func (a *analyzer) parseSendStmt(in *ast.SendStmt) {
	if n, ok := in.Value.(*ast.CallExpr); ok {
		a.parseCallExpr(n)
	}
}

func (a *analyzer) parseSwitchStmt(in *ast.SwitchStmt) {
	for _, x := range in.Body.List {
		if y, ok := x.(*ast.CaseClause); ok {
			for _, z := range y.Body {
				a.parseStmt(z)
			}
		}
	}
}

// parseTypeSwitchStmt parses
func (a *analyzer) parseTypeSwitchStmt(in *ast.TypeSwitchStmt) {
	if in.Body != nil {
		for _, x := range in.Body.List {
			if y, ok := x.(*ast.CaseClause); ok {
				for _, z := range y.Body {
					a.parseStmt(z)
				}
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
func (a *analyzer) parseStmt(in ast.Stmt) {
	switch e := in.(type) {
	case *ast.AssignStmt: // handles things like `e := Example{}` (with or without &)
		a.parseAssignStmt(e)
	// NOTE: even though RangeStmt/IfStmt/ForStmt are handled identically, Go will (rightfully) complain when trying
	// to use a multiple case statement (i.e. `case *ast.RangeStmt, *ast.IfStmt`), so we're doing it this way.
	case *ast.RangeStmt:
		for _, x := range e.Body.List {
			a.parseStmt(x)
		}
	case *ast.IfStmt:
		for _, x := range e.Body.List {
			a.parseStmt(x)
		}
	case *ast.ForStmt:
		for _, x := range e.Body.List {
			a.parseStmt(x)
		}
	case *ast.DeclStmt:
		a.parseDeclStmt(e)
	case *ast.ExprStmt:
		a.parseExprStmt(e)
	case *ast.DeferStmt:
		a.parseExpr(e.Call.Fun)
	case *ast.GoStmt:
		a.parseExpr(e.Call.Fun)
	case *ast.ReturnStmt:
		a.parseReturnStmt(e)
	case *ast.SelectStmt:
		a.parseSelectStmt(e)
	case *ast.SendStmt:
		a.parseSendStmt(e)
	case *ast.SwitchStmt:
		a.parseSwitchStmt(e)
	case *ast.TypeSwitchStmt:
		a.parseTypeSwitchStmt(e)
	}
}

func (a *analyzer) getDeclaredNames(in *ast.File) {
	for _, d := range in.Decls {
		if f, ok := d.(*ast.FuncDecl); ok {
			declPos := a.fileset.Position(f.Type.Func)
			functionName := a.parseFuncDecl(f)

			tf := BlanketFunc{
				Name:     functionName,
				Filename: declPos.Filename,
				DeclPos:  declPos,
			}

			if f.Body != nil {
				tf.RBracePos = a.fileset.Position(f.Body.Lbrace)
				tf.LBracePos = a.fileset.Position(f.Body.Rbrace)
			}
			a.declaredFuncInfo[functionName] = tf
		}
	}
}

func (a *analyzer) getCalledNames(in *ast.File) {
	for _, d := range in.Decls {
		switch n := d.(type) {
		case *ast.GenDecl:
			a.parseGenDecl(n)
		case *ast.FuncDecl:
			if _, ok := a.helperFunctionReturnMap[n.Name.Name]; !ok && n.Body != nil {
				for _, le := range n.Body.List {
					a.parseStmt(le)
				}
			}
		}
	}
}

func (a *analyzer) findHelperFuncs(in *ast.File) {
	for _, d := range in.Decls {
		if n, ok := d.(*ast.FuncDecl); ok {
			functionName := in.Name.Name
			if !strings.HasPrefix(functionName, "Test") {
				a.parseHelperFunction(n)
			}
		}
	}
}

func (a *analyzer) Analyze(analyzePackage string) (*BlanketReport, error) {
	gopath := os.Getenv("GOPATH")

	pkgDir := strings.Join([]string{gopath, "src", analyzePackage}, "/")
	if analyzePackage == "." {
		var err error
		pkgDir, err = os.Getwd()
		if err != nil {
			return nil, errors.Wrap(err, "getting current working directory")
		}
	}

	if a.debug {
		log.Printf("package directory: %s", pkgDir)
	}

	if _, err := os.Stat(pkgDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("packageDir doesn't exist: %s", pkgDir)
	}

	astPkg, err := parser.ParseDir(a.fileset, pkgDir, nil, parser.AllErrors)
	if err != nil {
		return nil, errors.Wrap(err, "parsing package directory")
	}

	if len(astPkg) == 0 || astPkg == nil {
		return nil, errors.New("no go files found!")
	}

	// find all helper funcs first so we have an idea of what they are.
	for _, pkg := range astPkg {
		for name, f := range pkg.Files {
			if strings.HasSuffix(name, "_test.go") {
				a.findHelperFuncs(f)
			}
		}
	}

	// find all the declared names
	for _, pkg := range astPkg {
		for name, f := range pkg.Files {
			if !strings.HasSuffix(name, "_test.go") {
				a.getDeclaredNames(f)
			}
		}
	}

	// find all the called names
	for _, pkg := range astPkg {
		for name, f := range pkg.Files {
			if strings.HasSuffix(name, "_test.go") {
				a.getCalledNames(f)
			}
		}
	}

	declaredFuncs := set.New()
	for _, f := range a.declaredFuncInfo {
		declaredFuncs.Add(f.Name)
	}

	for _, x := range set.StringSlice(set.Difference(a.calledFuncs, declaredFuncs)) {
		a.calledFuncs.Remove(x)
	}

	tr := &BlanketReport{
		DeclaredDetails: a.declaredFuncInfo,
		Declared:        declaredFuncs,
		Called:          a.calledFuncs,
	}
	return tr, nil
}

// NewAnalyzer creates a new instance of an Analyzer with some default values. It should be the only way an Analyzer is instantiated
func NewAnalyzer() *analyzer {
	return &analyzer{
		fileset:                 token.NewFileSet(),
		declaredFuncInfo:        map[string]BlanketFunc{},
		calledFuncs:             set.New("init"),
		helperFunctionReturnMap: map[string][]string{},
		nameToTypeMap:           map[string]string{},
	}
}

func (a *analyzer) GenerateDiffReport() *blanketOutput {
	if a.latestReport == nil {
		return nil
	}

	diff := set.StringSlice(set.Difference(a.latestReport.Declared, a.latestReport.Called))
	declaredFuncCount := a.latestReport.Declared.Size()
	calledFuncCount := a.latestReport.Called.Size()
	longestFunctionNameLength := 0
	missingFuncs := &blanketDetails{}
	for _, s := range diff {
		if utf8.RuneCountInString(s) > longestFunctionNameLength {
			longestFunctionNameLength = len(s)
		}
		*missingFuncs = append(*missingFuncs, a.declaredFuncInfo[s])
	}

	sort.Sort(missingFuncs)
	byFilename := map[string][]BlanketFunc{}
	for _, tf := range *missingFuncs {
		byFilename[tf.Filename] = append(byFilename[tf.Filename], tf)
	}
	score := float64(calledFuncCount) / float64(declaredFuncCount)

	return &blanketOutput{
		DeclaredCount:             declaredFuncCount,
		CalledCount:               calledFuncCount,
		Score:                     int(score * 100),
		Details:                   byFilename,
		LongestFunctionNameLength: longestFunctionNameLength,
	}
}
