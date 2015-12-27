package goscript

import (
	"fmt"
	"strings"

	. "go/ast"
	"go/token"
	"go/types"
	"go/importer"

	"github.com/nodirt/ast-rewrite"
)

var (
	errAstType = NewIdent("error")
	//errType = &types.NewNamed()
	nilExpr   = NewIdent("nil")
	panicExpr = NewIdent("panic")
)

type TransformationError struct {
	msg string
}

func (e *TransformationError) Error() string {
	return e.msg
}

type funcTransformer struct {
	FileSet *token.FileSet
	Info    *types.Info
	Pkg *types.Package

	typ        *FuncType
	body       *BlockStmt
	locals     *Scope
	outerScope *Scope
	errNames   []*Ident
	errors     []string
}

func (t *funcTransformer) err(format string, a ...interface{}) {
	t.errors = append(t.errors, fmt.Sprintf(format, a...))
}

func (t *funcTransformer) isUnique(name string) bool {
	scope := t.Info.Scopes[t.typ]
	if scope == nil {
		for n, s := range t.Info.Scopes {
			Print(t.FileSet, n)
			fmt.Printf("%s\n", s)
		}
		panic("bodyScope is nil")
	}
	if _, o := scope.LookupParent(name, token.NoPos); o != nil {
		return false
	}

	// bfs to check inner blocks
	queue := []*types.Scope{scope}
	for len(queue) > 0 {
		scope := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		if scope.Lookup(name) != nil {
			return false
		}
		for i := 0; i < scope.NumChildren(); i++ {
			queue = append(queue, scope.Child(i))
		}
	}
	return true
}

func (t *funcTransformer) allocErrVars(n int) {
	if len(t.errNames) < n {
		for i := len(t.errNames); len(t.errNames) < n; i++ {
			name := fmt.Sprintf("err%d", i)
			if t.isUnique(name) {
				t.errNames = append(t.errNames, NewIdent(name))
			}
		}
	}
}

func (t *funcTransformer) transform(name string, typ *FuncType, body *BlockStmt) error {
	t.typ = typ
	t.body = body
	t.errors = nil
	t.errNames = nil
	rewrite.Rewrite(body, t.rewrite)
	if t.errors != nil {
		return &TransformationError{
			fmt.Sprintf(
				"transformation of function %q failed:\n%s",
				name, strings.Join(t.errors, "\n")),
		}
	}
	if len(t.errNames) > 0 {
		errDecl := DeclStmt{&GenDecl{
			Tok: token.VAR,
			Specs: []Spec{
				&ValueSpec{
					Names: t.errNames,
					Type:  errAstType,
				},
			},
		}}
		body.List = append([]Stmt{&errDecl}, body.List...)
	}
	return nil
}

func (t *funcTransformer) decl(f *FuncDecl) error {
	return t.transform(f.Name.Name, f.Type, f.Body)
}

func (t *funcTransformer) lit(f *FuncLit) error {
	return t.transform("", f.Type, f.Body)
}

func (t *funcTransformer) rewrite(node Node) (n Node, continueRewriting bool) {
	switch n := node.(type) {
	case *AssignStmt:
		node = t.assign(n)
	case *ExprStmt:
		node = t.exprStmt(n)
	}
	return node, true
}

func (t *funcTransformer) genErrChecks(dest *rewrite.ExpandedBlockStmt, n int) {
	for i := 0; i < n; i++ {
		dest.List = append(dest.List, &IfStmt{
			Cond: &BinaryExpr{
				X:  t.errNames[i],
				Op: token.NEQ,
				Y:  nilExpr,
			},
			Body: &BlockStmt{
				List: []Stmt{
					&ExprStmt{
						&CallExpr{
							Fun:  panicExpr,
							Args: []Expr{t.errNames[i]},
						},
					},
				},
			},
		})
	}
}

func (t *funcTransformer) assign(assign *AssignStmt) Node {
	var blankErrors []int
	rhsTypes := make([]types.Type, len(assign.Lhs))
	switch len(assign.Rhs) {
	case len(assign.Lhs):
		for i, e := range assign.Rhs {
			rhsTypes[i] = t.Info.TypeOf(e)
		}
	case 1:
		switch typ := t.Info.TypeOf(assign.Rhs[0]).(type) {
		case *types.Tuple:
			for i := 0; i < typ.Len(); i++ {
				rhsTypes[i] = typ.At(i).Type()
			}
		default:
			panic(fmt.Sprintf("unexpected rhs type: %T", typ))
		}
	default:
		panic(fmt.Sprintf("unexpected len(rhs): %d", len(assign.Rhs)))
	}

	for i, expr := range assign.Lhs {
		if isBlank(expr) && isError(rhsTypes[i]) {
			blankErrors = append(blankErrors, i)
		}
	}
	if blankErrors == nil {
		return assign
	}

	t.allocErrVars(len(blankErrors))
	for i, j := range blankErrors {
		assign.Lhs[j] = t.errNames[i]
	}
	var result rewrite.ExpandedBlockStmt
	result.List = []Stmt{assign}
	t.genErrChecks(&result, len(blankErrors))
	return &result
}

func (t *funcTransformer) exprStmt(stmt *ExprStmt) Node {
	exprTypes := []types.Type{t.Info.TypeOf(stmt.X)}
	if tuple, ok := exprTypes[0].(*types.Tuple); ok {
		exprTypes = make([]types.Type, tuple.Len())
		for i := 0; i < tuple.Len(); i++ {
			exprTypes[i] = tuple.At(i).Type()
		}
	}

	var errorsIdx []int
	for i, typ := range exprTypes {
		if isError(typ) {
			errorsIdx = append(errorsIdx, i)
		}
	}
	if errorsIdx == nil {
		return stmt
	}

	t.allocErrVars(len(errorsIdx))
	assign := &AssignStmt{
		Lhs: make([]Expr, len(exprTypes)),
		Tok: token.ASSIGN,
		Rhs: []Expr{stmt.X},
	}
	for i, j := 0, 0; i < len(exprTypes); i++ {
		if errorsIdx[j] == i {
			assign.Lhs[i] = t.errNames[j]
			j++
		} else {
			assign.Lhs[i] = NewIdent("_")
		}
	}
	var result rewrite.ExpandedBlockStmt
	result.List = []Stmt{assign}
	t.genErrChecks(&result, len(errorsIdx))
	return &result
}

func Transform(files []*File, fset *token.FileSet) error {
	info := types.Info{
		Types:  make(map[Expr]types.TypeAndValue),
		Uses:   make(map[*Ident]types.Object),
		Defs:   make(map[*Ident]types.Object),
		Scopes: make(map[Node]*types.Scope),
	}


	conf := types.Config {
		Importer: importer.Default(),
	}
	pkg, err := conf.Check("", fset, files, &info)
	if err != nil {
		return err
	}

	for _, file := range files {
		for _, decl := range file.Decls {
			if fun, ok := decl.(*FuncDecl); ok {
				t := funcTransformer{Info: &info, FileSet: fset, Pkg: pkg}
				if err := t.decl(fun); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func isBlank(expr Expr) bool {
	if id, ok := expr.(*Ident); ok {
		return id.Name == "_"
	}
	return false
}

func isError(typ types.Type) bool {
	if named, ok := typ.(*types.Named); ok {
		o := named.Obj()
		return o != nil && o.Name() == "error" && o.Pkg() == nil
	}
	return false
}
