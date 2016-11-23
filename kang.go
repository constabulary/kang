package kang

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Context contains all build specific values.
type Context struct {
	GOOS, GOARCH string
	Workdir      string
	Pkgdir       string
	force        bool // always force build, even if not stale
	race         bool // build a -race enabled binary
	buildtags    []string
}

func (c *Context) isCrossCompile() bool { return false }

func (c *Context) Bindir() string { return c.Workdir }

// ctxString returns a string representation of the unique properties
// of the context.
func (c *Context) ctxString() string {
	v := []string{
		c.GOOS,
		c.GOARCH,
	}
	v = append(v, c.buildtags...)
	return strings.Join(v, "-")
}

// Package describes a set of Go files to be compiled.
type Package struct {
	*Context
	ImportPath string
	Dir        string
	GoFiles    []string
	Imports    []*Package
	standard   bool // is this part of the stdlib
	testScope  bool // is a test scoped packge
	Main       bool // this is a command
}

const debug = true

func debugf(format string, args ...interface{}) {
	if !debug {
		return
	}
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// IsStale returns true if the source pkg is considered to be stale with
// respect to its installed version.
func (pkg *Package) IsStale() bool {
	switch pkg.ImportPath {
	case "C", "unsafe":
		// synthetic packages are never stale
		return false
	}

	if !pkg.standard && pkg.force {
		return true
	}

	// tests are always stale, they are never installed
	if pkg.testScope {
		return true
	}

	// Package is stale if completely unbuilt.
	var built time.Time
	if fi, err := os.Stat(pkg.pkgpath()); err == nil {
		built = fi.ModTime()
	}

	if built.IsZero() {
		debugf("%s is missing", pkg.pkgpath())
		return true
	}

	olderThan := func(file string) bool {
		fi, err := os.Stat(file)
		return err != nil || fi.ModTime().After(built)
	}

	newerThan := func(file string) bool {
		fi, err := os.Stat(file)
		return err != nil || fi.ModTime().Before(built)
	}

	if pkg.standard && !pkg.isCrossCompile() {
		// if this is a standard lib package, and we are not cross compiling
		// then assume the package is up to date. This also works around
		// golang/go#13769.
		return false
	}

	// Package is stale if a dependency is newer.
	for _, p := range pkg.Imports {
		if p.ImportPath == "C" || p.ImportPath == "unsafe" {
			continue // ignore stale imports of synthetic packages
		}
		if olderThan(p.pkgpath()) {
			debugf("%s is older than %s", pkg.pkgpath(), p.pkgpath())
			return true
		}
	}

	// if the main package is up to date but _newer_ than the binary (which
	// could have been removed), then consider it stale.
	if pkg.Main && newerThan(pkg.Binfile()) {
		debugf("%s is newer than %s", pkg.pkgpath(), pkg.Binfile())
		return true
	}

	for _, src := range pkg.files() {
		if olderThan(filepath.Join(pkg.Dir, src)) {
			debugf("%s is older than %s", pkg.pkgpath(), filepath.Join(pkg.Dir, src))
			return true
		}
	}

	return false
}

// files returns all source files in scope
func (p *Package) files() []string {
	return stringList(p.GoFiles)
}

// pkgpath returns the destination for object cached for this Package.
func (pkg *Package) pkgpath() string {
	importpath := filepath.FromSlash(pkg.ImportPath) + ".a"
	switch {
	case pkg.isCrossCompile():
		return filepath.Join(pkg.Pkgdir, importpath)
	case pkg.standard && pkg.race:
		// race enabled standard lib
		return filepath.Join(runtime.GOROOT(), "pkg", pkg.GOOS+"_"+pkg.GOARCH+"_race", importpath)
	case pkg.standard:
		// standard lib
		return filepath.Join(runtime.GOROOT(), "pkg", pkg.GOOS+"_"+pkg.GOARCH, importpath)
	default:
		return filepath.Join(pkg.Pkgdir, importpath)
	}
}

// Binfile returns the destination of the compiled target of this command.
func (pkg *Package) Binfile() string {
	// TODO(dfc) should have a check for package main, or should be merged in to objfile.
	target := filepath.Join(pkg.Bindir(), pkg.binname())
	if pkg.testScope {
		target = filepath.Join(pkg.Workdir, filepath.FromSlash(pkg.ImportPath), "_test", pkg.binname())
	}

	// if this is a cross compile or GOOS/GOARCH are both defined or there are build tags, add ctxString.
	if pkg.isCrossCompile() || (os.Getenv("GOOS") != "" && os.Getenv("GOARCH") != "") {
		target += "-" + pkg.ctxString()
	} else if len(pkg.buildtags) > 0 {
		target += "-" + strings.Join(pkg.buildtags, "-")
	}

	if pkg.GOOS == "windows" {
		target += ".exe"
	}
	return target
}

func (pkg *Package) binname() string {
	switch {
	case pkg.testScope:
		return filepath.FromSlash(pkg.ImportPath) + ".test"
	case pkg.Main:
		return filepath.Base(filepath.FromSlash(pkg.ImportPath))
	default:
		panic("binname called with non main package: " + pkg.ImportPath)
	}
}

func stringList(args ...[]string) []string {
	var l []string
	for _, arg := range args {
		l = append(l, arg...)
	}
	return l
}
