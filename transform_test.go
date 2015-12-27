package goscript

import (
	"bytes"
	"go/ast"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go/format"
	"go/parser"
	"go/token"
	"go/types"
)

const (
	TEST_DATA_DIR = "testdata"
	IN_SUFFIX     = ".in.go"
	OUT_SUFFIX    = ".out.go"
)

func TestIsBlank(t *testing.T) {
	fun := parseFunc(t, `func () {
		_ = 1
	}`)
	_ = 1
	assign := fun.Body.List[0].(*ast.AssignStmt)
	if !isBlank(assign.Lhs[0]) {
		t.FailNow()
	}
}

func parseFunc(t *testing.T, fun string) *ast.FuncLit {
	expr, err := parser.ParseExpr(fun)
	if err != nil {
		t.Fatal(err)
	}
	return expr.(*ast.FuncLit)
}

func TestFuncTransformer(t *testing.T) {
	testdata, err := os.Open(TEST_DATA_DIR)
	if err != nil {
		t.Fatal(err)
	}
	allFiles, err := testdata.Readdir(0)
	if err != nil {
		t.Fatal(err)
	}
	for _, inFileInfo := range allFiles {
		inFile := filepath.Join(TEST_DATA_DIR, inFileInfo.Name())
		if !strings.HasSuffix(inFile, IN_SUFFIX) {
			continue
		}
		input, err := ioutil.ReadFile(inFile)
		if err != nil {
			t.Fatal(err)
		}
		testName := strings.TrimSuffix(inFile, IN_SUFFIX)
		outFile := testName + OUT_SUFFIX
		expected, err := ioutil.ReadFile(outFile)
		if err != nil {
			t.Fatal(err)
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "test.go", input, 0)
		if err != nil {
			t.Fatal(err)
		}
		if err := Transform([]*ast.File{file}, fset); err != nil {
			t.Fatal(err)
		}
		var buf bytes.Buffer
		if err := format.Node(&buf, fset, file); err != nil {
			t.Fatal(err)
		}
		actual := buf.Bytes()
		if !bytes.Equal(expected, actual) {
			t.Logf("want \n%s; got \n%s", expected, actual)
			i := 0
			for ; i < len(expected) && i < len(actual) && expected[i] == actual[i]; i++ {
			}
			t.Logf("first mismatching runes: %q != %q", expected[i:], actual[i:])
			t.Fatalf("%s failed", testName)
		}
	}
}

func TestX(t *testing.T) {
	fset := token.NewFileSet()
	src := `
	package x

	func f(g func() (int, error)) {
		_, _ = g()
}`
	file, err := parser.ParseFile(fset, "helloWorld.go", src, 0)
	if err != nil {
		panic(err)
	}

	info := types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Uses:  make(map[*ast.Ident]types.Object),
		Defs:  make(map[*ast.Ident]types.Object),
	}
	var cfg types.Config
	_, err = cfg.Check("a", fset, []*ast.File{file}, &info)
	if err != nil {
		panic(err)
	}
	for e, typ := range info.Types {
		if typ.Type.String() == "error" {
			t.Logf("%T %s", typ.Type, typ.Type)
			if named, ok := typ.Type.(*types.Named); ok {
				t.Log(named.Obj().Id())
				t.Log(named.Obj().Pkg())
				t.Logf("%s - %#v", e, typ.Type)
			}
		}
	}
}
