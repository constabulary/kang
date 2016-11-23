package kang

// Context contains all build specific values.
type Context struct {
	GOOS, GOARCH string
	Workdir      string
	Pkgdir       string
}

// Package describes a set of Go files to be compiled.
type Package struct {
	*Context
	ImportPath string
	GoFiles    []string
	Deps       []*Package
}
