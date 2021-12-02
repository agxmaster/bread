all: goclean gotest gorace

gogen:
	go generate -v git.qutoutiao.net/golib/redis/gen

goclean:
	go clean -v -r ./...
	go mod tidy -v

gotest:
	go test -v -timeout 10s -race git.qutoutiao.net/golib/redis
	go test -v -timeout 10s git.qutoutiao.net/golib/redis

gorace: goclean
	go test -v -timeout 10s -race git.qutoutiao.net/gopher/qulibs/gorm
	go test -v -timeout 10s -race git.qutoutiao.net/gopher/qulibs/mongo
	go test -v -timeout 10s -race git.qutoutiao.net/gopher/qulibs/redis

gobench: goclean
	go test -run=none -bench=.


