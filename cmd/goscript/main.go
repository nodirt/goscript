package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"go/format"

	"github.com/nodirt/goscript"
	"strings"
)

func usage() {
	fmt.Fprintln(os.Stderr, "usage: goscript run <.go files> [--] [args]")
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	subcommand := os.Args[1]
	if subcommand != "run" {
		usage()
	}


	var fileNames, args []string
	fileNames = os.Args[2:]
	for i, s := range fileNames {
		if !strings.HasSuffix(s, ".go") {
			argStart := i
			if s == "--" {
				argStart++
			}
			args = fileNames[argStart:]
			fileNames = fileNames[:i]
			break
		}
	}
	if err := Run(fileNames, args); err != nil {
		exitCode := 2
		if exit, ok := err.(*exec.ExitError); ok {
			if sys, ok := exit.Sys().(syscall.WaitStatus); ok {
				exitCode = sys.ExitStatus()
			}
		} else {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(exitCode)
	}
}

func Run(fileNames, args []string) error {
	if len(fileNames) == 0 {
		return fmt.Errorf(".go files not specified")
	}
	parentDir := filepath.Dir(fileNames[0])
	fset := token.NewFileSet()
	files := make([]*ast.File, len(fileNames))
	for i, filename := range fileNames {
		dir := filepath.Dir(filename)
		if dir != parentDir {
			return fmt.Errorf("files belong to different directories")
		}
		var err error
		files[i], err = parser.ParseFile(fset, filename, nil, 0)
		if err != nil {
			return err
		}
	}

	if err := goscript.Transform(files, fset); err != nil {
		return err
	}

	tmpDir, err := ioutil.TempDir("", "goscript")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)
	finalFileNames := make([]string, len(fileNames))
	for i, filename := range fileNames {
		finalFileNames[i] = filepath.Join(tmpDir, filepath.Base(filename))
		dest, err := os.Create(finalFileNames[i])
		if err != nil {
			return err
		}
		err = format.Node(dest, fset, files[i])
		dest.Close()
		if err != nil {
			return err
		}
	}

	cmd := exec.Command("go", append(append([]string{"run"}, finalFileNames...), args...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
