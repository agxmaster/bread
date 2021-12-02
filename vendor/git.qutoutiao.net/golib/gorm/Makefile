all: goclean gotest gorace

gogen:
	go generate -v git.qutoutiao.net/golib/gorm/gen

goclean:
	go clean -v -r ./...
	go mod tidy -v

gotest:
	go test -v -timeout 10s -race git.qutoutiao.net/golib/gorm
	go test -v -timeout 10s git.qutoutiao.net/golib/gorm

gorace: goclean
	go test -v -timeout 10s -race git.qutoutiao.net/gopher/qulibs/gorm

gobench: goclean
	go test -run=none -bench=.


