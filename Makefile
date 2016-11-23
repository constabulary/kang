kang-boostrap: .kang/bootstrap/github.com/constabulary/kang.a .kang/bootstrap/github.com/constabulary/kang/cmd/kang.a

.kang/bootstrap/github.com/constabulary/kang/cmd/kang.a: .kang/bootstrap/github.com/constabulary/kang/cmd .kang/bootstrap/github.com/constabulary/kang.a 

.kang/bootstrap/github.com/constabulary/kang.a: .kang/bootstrap/github.com/constabulary

.kang/bootstrap/github.com/constabulary:
	mkdir -p $@

.kang/bootstrap/github.com/constabulary/kang/cmd:
	mkdir -p $@

clean:
	rm -rf .kang/ kang-bootstrap
