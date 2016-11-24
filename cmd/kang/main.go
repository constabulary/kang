package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"github.com/constabulary/kang"
)

func check(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %+v\n", err)
		os.Exit(1)
	}
}

func main() {
	kangfile := flag.String("f", ".kangfile", "path to .kangfile")
	flag.Parse()

	workdir, err := ioutil.TempDir("", "kang")
	check(err)

	*kangfile, err = filepath.Abs(*kangfile)
	check(err)

	rootdir := filepath.Dir(*kangfile)
	pkgdir := filepath.Join(rootdir, ".kang", "pkg")

	ctx := &kang.Context{
		GOOS:    runtime.GOOS,
		GOARCH:  runtime.GOARCH,
		Workdir: workdir,
		Pkgdir:  pkgdir,
		Bindir:  rootdir,
	}

	pkg := &kang.Package{
		Context:    ctx,
		ImportPath: "github.com/constabulary/kang",
		Dir:        rootdir,
		GoFiles:    []string{"kang.go"},
	}
	pkg.NotStale = !pkg.IsStale()

	pkg = &kang.Package{
		Context:    ctx,
		ImportPath: "github.com/constabulary/cmd/kang",
		Main:       true,
		Dir:        filepath.Join(rootdir, "cmd", "kang"),
		GoFiles:    []string{"main.go"},
		Imports:    []*kang.Package{pkg},
	}

	pkg.NotStale = !pkg.IsStale()

	build(pkg)
}

func build(pkgs ...*kang.Package) {
	targets := make(map[string]func() error)

	fn, err := buildPackages(targets, pkgs)
	check(err)
	check(fn())
}

func buildPackages(targets map[string]func() error, pkgs []*kang.Package) (func() error, error) {
	var deps []func() error
	for _, pkg := range pkgs {
		fn, err := buildPackage(targets, pkg)
		check(err)
		deps = append(deps, fn)
	}
	return func() error {
		for _, fn := range deps {
			if err := fn(); err != nil {
				return err
			}
		}
		return nil
	}, nil
}

func buildPackage(targets map[string]func() error, pkg *kang.Package) (func() error, error) {

	// if this action is already present in the map, return it
	// rather than creating a new action.
	if fn, ok := targets[pkg.ImportPath]; ok {
		return fn, nil
	}

	// step 0. are we stale ?
	// if this package is not stale, then by definition none of its
	// dependencies are stale, so ignore this whole tree.
	if pkg.NotStale {
		return func() error {
			fmt.Println(pkg.ImportPath, "is up to date")
			return nil
		}, nil
	}

	// step 1. build dependencies
	var deps []func() error
	for _, pkg := range pkg.Imports {
		fn, err := buildPackage(targets, pkg)
		if err != nil {
			return nil, err
		}
		deps = append(deps, fn)
	}

	// step 2. build this package
	build := func() error {
		for _, dep := range deps {
			if err := dep(); err != nil {
				return err
			}
		}
		if err := kang.Compile(pkg); err != nil {
			return err
		}
		if !pkg.Main {
			return nil // we're done
		}
		return kang.Link(pkg)
	}

	// record the final action as the action that represents
	// building this package.
	targets[pkg.ImportPath] = build

	return build, nil
}
