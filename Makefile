build: .kang/kang-bootstrap
	$^ build

.kang/kang-bootstrap: .kang/bootstrap/github.com/constabulary/kang/cmd/kang.a
	go tool link -o $@ -L .kang/bootstrap -w -extld=gcc -buildmode=exe $^

.kang/bootstrap/github.com/constabulary/kang/cmd/kang.a: .kang/bootstrap/github.com/constabulary/kang.a  cmd/kang/main.go
	mkdir -p .kang/bootstrap/github.com/constabulary/kang/cmd/
	go tool compile -o $@ -p kang -complete -I .kang/bootstrap -pack cmd/kang/main.go

.kang/bootstrap/github.com/constabulary/kang.a: kang.go
	mkdir -p .kang/bootstrap/github.com/constabulary
	go tool compile -o $@ -p github.com/constabulary/kang -complete -I .kang/bootstrap -pack kang.go

clean:
	rm -rf .kang/
