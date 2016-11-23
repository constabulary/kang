package kang

// Context contains all build specific values.
type Context struct {
	GOOS, GOARCH string
	Workdir      string
	Pkgdir       string
}
