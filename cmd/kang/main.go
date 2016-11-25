package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/constabulary/kang"
)

func check(err error) {
	if err != nil {
		fatal(err)
		os.Exit(1)
	}
}

func fatal(arg interface{}, args ...interface{}) {
	fmt.Fprint(os.Stderr, "fatal: ", arg)
	fmt.Fprintln(os.Stderr, args...)
	os.Exit(1)
}

func main() {
	flag.Parse()
	f, err := findkangfile(cwd())
	check(err)

	f, err = filepath.Abs(f)
	check(err)

	fmt.Println("Using", f)

	kf, err := ParseFile(f)
	check(err)

	prefix, ok := kf["project"]["prefix"]
	if prefix == "" || !ok {
		fatal("project prefix missing from .kangfile")
	}

	workdir, err := ioutil.TempDir("", "kang")
	check(err)

	rootdir := filepath.Dir(f)
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
		ImportPath: prefix,
		Dir:        rootdir,
		GoFiles:    []string{"kang.go"},
	}
	pkg.NotStale = !pkg.IsStale()

	main := &kang.Package{
		Context:    ctx,
		ImportPath: path.Join(prefix, "cmd", "kang"),
		Main:       true,
		Dir:        filepath.Join(rootdir, "cmd", "kang"),
		GoFiles:    []string{"main.go", "kangfile.go"},
		Imports:    []*kang.Package{pkg},
	}

	main.NotStale = !main.IsStale()

	build(main)
}

func cwd() string {
	wd, err := os.Getwd()
	check(err)
	return wd
}

// findkangfile returns the location of the closest .kangfile
// relative to the dir provided. If no .kangfile is found, an
// error is returned.
func findkangfile(dir string) (string, error) {
	orig := dir
	for {
		path := filepath.Join(dir, ".kangfile")
		fi, err := os.Stat(path)
		if err == nil && !fi.IsDir() {
			return path, nil
		}
		if err != nil && !os.IsNotExist(err) {
			check(err)
		}
		d := filepath.Dir(dir)
		if d == dir {
			// got to the root directory without
			return "", fmt.Errorf("could not locate .kangfile in %s", orig)
		}
		dir = d
	}
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
