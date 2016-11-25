build: kang
	./$^ build
	
kang: .kang/kang-bootstrap
	$^ build

.kang/kang-bootstrap: .kang/bootstrap/github.com/constabulary/kang/cmd/kang.a
	go tool link -o $@ -L .kang/bootstrap -w -extld=gcc -buildmode=exe $^

.kang/bootstrap/github.com/constabulary/kang/cmd/kang.a: \
	.kang/bootstrap/github.com/constabulary/kang.a \
	.kang/bootstrap/github.com/constabulary/kang/cmd/kang/internal/kangfile.a \
	cmd/kang/main.go
	mkdir -p .kang/bootstrap/github.com/constabulary/kang/cmd/
	go tool compile -o $@ -p github.com/constabular/cmd/kang -complete -I .kang/bootstrap -pack cmd/kang/main.go

.kang/bootstrap/github.com/constabulary/kang.a: kang.go
	mkdir -p .kang/bootstrap/github.com/constabulary
	go tool compile -o $@ -p github.com/constabulary/kang -complete -I .kang/bootstrap -pack $^

.kang/bootstrap/github.com/constabulary/kang/cmd/kang/internal/kangfile.a: cmd/kang/internal/kangfile/kangfile.go
	mkdir -p .kang/bootstrap/github.com/constabulary/kang/cmd/kang/internal/
	go tool compile -o $@ -p github.com/constabulary/kang/cmd/kang/internal/kangfile -complete -I .kang/bootstrap -pack $^

clean:
	rm -rf .kang/ kang
