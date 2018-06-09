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

func (a *analyzer) parseExpr(in ast.Expr, nameToTypeMap map[string]string, helperFunctionReturnMap map[string][]string, out *set.Set) {
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
		a.parseFuncLit(f, nameToTypeMap, helperFunctionReturnMap, out)
	}
}

func (a *analyzer) parseCallExpr(in *ast.CallExpr, nameToTypeMap map[string]string, helperFunctionReturnMap map[string][]string, out *set.Set) {
	for _, arg := range in.Args {
		switch at := arg.(type) {
		// case *ast.Ident:
		case *ast.CallExpr:
			a.parseCallExpr(at, nameToTypeMap, helperFunctionReturnMap, out)
		case *ast.FuncLit:
			a.parseFuncLit(at, nameToTypeMap, helperFunctionReturnMap, out)
		}
	}
	a.parseExpr(in.Fun, nameToTypeMap, helperFunctionReturnMap, out)
}

// parseUnaryExpr parses Unary expressions. From the go/ast docs:
//      A UnaryExpr node represents a unary expression. Unary "*" expressions are represented via StarExpr nodes.
// (handles declarations like `callExpr := &ast.UnaryExpr{}` or `callExpr := ast.UnaryExpr{}`)
func (a *analyzer) parseUnaryExpr(in *ast.UnaryExpr, varName string, nameToTypeMap map[string]string, helperFunctionReturnMap map[string][]string, out *set.Set) {
	if cl, ok := in.X.(*ast.CompositeLit); ok {
		for _, e := range cl.Elts {
			switch et := e.(type) {
			case *ast.CallExpr:
				a.parseExpr(et.Fun, nameToTypeMap, helperFunctionReturnMap, out)
			case *ast.KeyValueExpr:
				if vt, ok := et.Value.(*ast.CallExpr); ok {
					a.parseCallExpr(vt, nameToTypeMap, helperFunctionReturnMap, out)
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
func (a *analyzer) parseDeclStmt(in *ast.DeclStmt, nameToTypeMap map[string]string) {
	if gd, ok := in.Decl.(*ast.GenDecl); ok {
		if len(gd.Specs) > 0 {
			if s, ok := gd.Specs[0].(*ast.ValueSpec); ok {
				if len(s.Names) > 0 {
					varName := s.Names[0].Name
					switch t := s.Type.(type) {
					case *ast.Ident:
						nameToTypeMap[varName] = t.Name
					case *ast.SelectorExpr:
						nameToTypeMap[varName] = t.Sel.Name
					}
				}
			}
		}
	}
}

// parseExprStmt parses expression statements. From the go/ast docs:
// 		An ExprStmt node represents a (stand-alone) expression in a statement list.
func (a *analyzer) parseExprStmt(in *ast.ExprStmt, nameToTypeMap map[string]string, helperFunctionReturnMap map[string][]string, out *set.Set) {
	if c, ok := in.X.(*ast.CallExpr); ok {
		a.parseCallExpr(c, nameToTypeMap, helperFunctionReturnMap, out)
	}
}

func (a *analyzer) parseCompositeLit(in *ast.CompositeLit, varName string, nameToTypeMap map[string]string, helperFunctionReturnMap map[string][]string, out *set.Set) {
	for _, e := range in.Elts {
		if et, ok := e.(*ast.CallExpr); ok {
			a.parseExpr(et.Fun, nameToTypeMap, helperFunctionReturnMap, out)
		}
	}

	switch t := in.Type.(type) {
	case *ast.Ident:
		nameToTypeMap[varName] = t.Name
	case *ast.SelectorExpr:
		nameToTypeMap[varName] = t.Sel.Name
	}
}

// parseGenDecl handles GenDecl nodes. From the go/ast docs:
//     A GenDecl node (generic declaration node) represents an import, constant, type or variable declaration.
func (a *analyzer) parseGenDecl(in *ast.GenDecl, nameToTypeMap map[string]string) {
	for _, spec := range in.Specs {
		if global, ok := spec.(*ast.ValueSpec); ok {
			if len(global.Names) > 0 {
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
func (a *analyzer) parseAssignStmt(in *ast.AssignStmt, nameToTypeMap map[string]string, helperFunctionReturnMap map[string][]string, out *set.Set) {
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
			a.parseFuncLit(t, nameToTypeMap, helperFunctionReturnMap, out)
		case *ast.UnaryExpr:
			a.parseUnaryExpr(t, leftHandSide[j], nameToTypeMap, helperFunctionReturnMap, out)
		case *ast.CompositeLit:
			if len(leftHandSide) > j {
				a.parseCompositeLit(t, leftHandSide[j], nameToTypeMap, helperFunctionReturnMap, out)
			} else {
				a.parseCompositeLit(t, "", nameToTypeMap, helperFunctionReturnMap, out)
			}
		case *ast.CallExpr:
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

			a.parseCallExpr(t, nameToTypeMap, helperFunctionReturnMap, out)
		}
	}
}

func (a *analyzer) parseHelperSelectorExpr(in *ast.SelectorExpr, functionName string, helperFunctionReturnMap map[string][]string) {
	if pkg, ok := in.X.(*ast.Ident); ok {
		pkgName := pkg.Name
		pkgStruct := in.Sel.Name
		helperFunctionReturnMap[functionName] = append(helperFunctionReturnMap[functionName], fmt.Sprintf("%s.%s", pkgName, pkgStruct))
	}
}

func (a *analyzer) parseHelperFunction(in *ast.FuncDecl, helperFunctionReturnMap map[string][]string, out *set.Set) {
	functionName := in.Name.Name
	if in.Type.Results != nil {
		for _, r := range in.Type.Results.List {
			switch rt := r.Type.(type) {
			case *ast.SelectorExpr:
				a.parseHelperSelectorExpr(rt, functionName, helperFunctionReturnMap)
			case *ast.StarExpr:
				switch x := rt.X.(type) {
				case *ast.Ident:
					helperFunctionReturnMap[functionName] = append(helperFunctionReturnMap[functionName], x.Name)
				case *ast.SelectorExpr:
					a.parseHelperSelectorExpr(x, functionName, helperFunctionReturnMap)
				}
			case *ast.Ident:
				helperFunctionReturnMap[functionName] = append(helperFunctionReturnMap[functionName], rt.Name)
			}
		}
	}
}

// parseFuncLit parses a function literal. From the go/ast docs:
// 		A FuncLit node represents a function literal.
// FuncLits have bodies that we basically need to explore the same way that we explore a normal function.
func (a *analyzer) parseFuncLit(in *ast.FuncLit, nameToTypeMap map[string]string, helperFunctionReturnMap map[string][]string, out *set.Set) {
	for _, le := range in.Body.List {
		a.parseStmt(le, nameToTypeMap, helperFunctionReturnMap, out)
	}
}

func (a *analyzer) parseReturnStmt(in *ast.ReturnStmt, nameToTypeMap map[string]string, helperFunctionReturnMap map[string][]string, out *set.Set) {
	for _, x := range in.Results {
		if y, ok := x.(*ast.CallExpr); ok {
			a.parseExpr(y.Fun, nameToTypeMap, helperFunctionReturnMap, out)
		}
	}
}

func (a *analyzer) parseSelectStmt(in *ast.SelectStmt, nameToTypeMap map[string]string, helperFunctionReturnMap map[string][]string, out *set.Set) {
	for _, x := range in.Body.List {
		if y, ok := x.(*ast.CommClause); ok {
			for _, z := range y.Body {
				a.parseStmt(z, nameToTypeMap, helperFunctionReturnMap, out)
			}
		}
	}
}

// parseSendStmt parses a send statement. (<-)
func (a *analyzer) parseSendStmt(in *ast.SendStmt, nameToTypeMap map[string]string, helperFunctionReturnMap map[string][]string, out *set.Set) {
	if n, ok := in.Value.(*ast.CallExpr); ok {
		a.parseCallExpr(n, nameToTypeMap, helperFunctionReturnMap, out)
	}
}

func (a *analyzer) parseSwitchStmt(in *ast.SwitchStmt, nameToTypeMap map[string]string, helperFunctionReturnMap map[string][]string, out *set.Set) {
	for _, x := range in.Body.List {
		if y, ok := x.(*ast.CaseClause); ok {
			for _, z := range y.Body {
				a.parseStmt(z, nameToTypeMap, helperFunctionReturnMap, out)
			}
		}
	}
}

// parseTypeSwitchStmt parses
func (a *analyzer) parseTypeSwitchStmt(in *ast.TypeSwitchStmt, nameToTypeMap map[string]string, helperFunctionReturnMap map[string][]string, out *set.Set) {
	if in.Body != nil {
		for _, x := range in.Body.List {
			if y, ok := x.(*ast.CaseClause); ok {
				for _, z := range y.Body {
					a.parseStmt(z, nameToTypeMap, helperFunctionReturnMap, out)
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
func (a *analyzer) parseStmt(in ast.Stmt, nameToTypeMap map[string]string, helperFunctionReturnMap map[string][]string, out *set.Set) {
	switch e := in.(type) {
	case *ast.AssignStmt: // handles things like `e := Example{}` (with or without &)
		a.parseAssignStmt(e, nameToTypeMap, helperFunctionReturnMap, out)
	// NOTE: even though RangeStmt/IfStmt/ForStmt are handled identically, Go will (rightfully) complain when trying
	// to use a multiple case statement (i.e. `case *ast.RangeStmt, *ast.IfStmt`), so we're doing it this way.
	case *ast.RangeStmt:
		for _, x := range e.Body.List {
			a.parseStmt(x, nameToTypeMap, helperFunctionReturnMap, out)
		}
	case *ast.IfStmt:
		for _, x := range e.Body.List {
			a.parseStmt(x, nameToTypeMap, helperFunctionReturnMap, out)
		}
	case *ast.ForStmt:
		for _, x := range e.Body.List {
			a.parseStmt(x, nameToTypeMap, helperFunctionReturnMap, out)
		}
	case *ast.DeclStmt:
		a.parseDeclStmt(e, nameToTypeMap)
	case *ast.ExprStmt:
		a.parseExprStmt(e, nameToTypeMap, helperFunctionReturnMap, out)
	case *ast.DeferStmt:
		a.parseExpr(e.Call.Fun, nameToTypeMap, helperFunctionReturnMap, out)
	case *ast.GoStmt:
		a.parseExpr(e.Call.Fun, nameToTypeMap, helperFunctionReturnMap, out)
	case *ast.ReturnStmt:
		a.parseReturnStmt(e, nameToTypeMap, helperFunctionReturnMap, out)
	case *ast.SelectStmt:
		a.parseSelectStmt(e, nameToTypeMap, helperFunctionReturnMap, out)
	case *ast.SendStmt:
		a.parseSendStmt(e, nameToTypeMap, helperFunctionReturnMap, out)
	case *ast.SwitchStmt:
		a.parseSwitchStmt(e, nameToTypeMap, helperFunctionReturnMap, out)
	case *ast.TypeSwitchStmt:
		a.parseTypeSwitchStmt(e, nameToTypeMap, helperFunctionReturnMap, out)
	}
}

func (a *analyzer) getDeclaredNames(in *ast.File, fileset *token.FileSet, declaredFuncDetails map[string]BlanketFunc) {
	for _, d := range in.Decls {
		if f, ok := d.(*ast.FuncDecl); ok {
			declPos := fileset.Position(f.Type.Func)
			functionName := a.parseFuncDecl(f)

			tf := BlanketFunc{
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

func (a *analyzer) getCalledNames(in *ast.File, nameToTypeMap map[string]string, helperFunctionReturnMap map[string][]string, out *set.Set) {
	for _, d := range in.Decls {
		switch n := d.(type) {
		case *ast.GenDecl:
			a.parseGenDecl(n, nameToTypeMap)
		case *ast.FuncDecl:
			if _, ok := helperFunctionReturnMap[n.Name.Name]; !ok && n.Body != nil {
				for _, le := range n.Body.List {
					a.parseStmt(le, nameToTypeMap, helperFunctionReturnMap, out)
				}
			}
		}
	}
}

func (a *analyzer) findHelperFuncs(in *ast.File, helperFunctionReturnMap map[string][]string, out *set.Set) {
	for _, d := range in.Decls {
		if n, ok := d.(*ast.FuncDecl); ok {
			functionName := in.Name.Name
			if !strings.HasPrefix(functionName, "Test") {
				a.parseHelperFunction(n, helperFunctionReturnMap, out)
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

	_, err := os.Stat(pkgDir)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("packageDir doesn't exist: %s", pkgDir)
	}

	astPkg, err := parser.ParseDir(a.fileset, pkgDir, nil, parser.AllErrors)
	if err != nil {
		return nil, errors.Wrap(err, "parsing package directory")
	}

	if len(astPkg) == 0 || astPkg == nil {
		return nil, errors.New("no go files found!")
	}

	declaredFuncInfo := map[string]BlanketFunc{}
	calledFuncs := set.New("init")
	helperFunctionReturnMap := map[string][]string{}
	nameToTypeMap := map[string]string{}

	// find all helper funcs first so we have an idea of what they are.
	for _, pkg := range astPkg {
		for name, f := range pkg.Files {
			if strings.HasSuffix(name, "_test.go") {
				a.findHelperFuncs(f, helperFunctionReturnMap, calledFuncs)
			}
		}
	}

	// find all the declared names
	for _, pkg := range astPkg {
		for name, f := range pkg.Files {
			if !strings.HasSuffix(name, "_test.go") {
				a.getDeclaredNames(f, a.fileset, declaredFuncInfo)
			}
		}
	}

	// find all the called names
	for _, pkg := range astPkg {
		for name, f := range pkg.Files {
			if strings.HasSuffix(name, "_test.go") {
				a.getCalledNames(f, nameToTypeMap, helperFunctionReturnMap, calledFuncs)
			}
		}
	}

	declaredFuncs := set.New()
	for _, f := range declaredFuncInfo {
		declaredFuncs.Add(f.Name)
	}

	for _, x := range set.StringSlice(set.Difference(calledFuncs, declaredFuncs)) {
		calledFuncs.Remove(x)
	}

	tr := &BlanketReport{
		DeclaredDetails: declaredFuncInfo,
		Declared:        declaredFuncs,
		Called:          calledFuncs,
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
