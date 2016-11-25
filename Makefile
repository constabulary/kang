build: kang
	./$^ build
	
kang: .kang/kang-bootstrap
	$^ build

.kang/kang-bootstrap: .kang/bootstrap/github.com/constabulary/kang/cmd/kang.a
	go tool link -o $@ -L .kang/bootstrap -w -extld=gcc -buildmode=exe $^

.kang/bootstrap/github.com/constabulary/kang/cmd/kang.a: .kang/bootstrap/github.com/constabulary/kang.a cmd/kang/main.go cmd/kang/kangfile.go cmd/kang/stdlib.go
	mkdir -p .kang/bootstrap/github.com/constabulary/kang/cmd/
	go tool compile -o $@ -p github.com/constabular/cmd/kang -complete -I .kang/bootstrap -pack cmd/kang/main.go cmd/kang/kangfile.go cmd/kang/stdlib.go

.kang/bootstrap/github.com/constabulary/kang.a: kang.go
	mkdir -p .kang/bootstrap/github.com/constabulary
	go tool compile -o $@ -p github.com/constabulary/kang -complete -I .kang/bootstrap -pack $^

clean:
	rm -rf .kang/ kang
