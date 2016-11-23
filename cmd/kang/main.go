package main

import "flag"

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

	kangfile, err = filepath.Abs(kangfile)
	check(err)

	rootdir := filepath.Base(kangfile)
	pkgdir := filepath.Join(rootdir, ".kang", "pkg")

	ctx := &kang.Context {
		GOOS: runtime.GOOS,
		GOARCH: runtime.GOARCH,
		Workdir: workdir,
		Pkgdir: pkgdir,
	}

	pkgs := []*Package {{
		Contet: ctx,
		ImportPath: "github.com/constabulary/kang",
		GoFiles: []string{"kang.go"},
	}}
}

}
