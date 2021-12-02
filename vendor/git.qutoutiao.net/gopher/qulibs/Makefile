all: goclean gotest gorace

gogen:
	go generate -v git.qutoutiao.net/gopher/qulibs/redis/gen

goclean:
	go clean -v -r ./...
	go mod tidy -v

gotest:
	go test -v -timeout 10s -race git.qutoutiao.net/gopher/qulibs/gorm
	go test -v -timeout 10s git.qutoutiao.net/gopher/qulibs/gorm
	go test -v -timeout 10s -race git.qutoutiao.net/gopher/qulibs/mongo
	go test -v -timeout 10s git.qutoutiao.net/gopher/qulibs/mongo
	go test -v -timeout 10s -race git.qutoutiao.net/gopher/qulibs/redis
	go test -v -timeout 10s git.qutoutiao.net/gopher/qulibs/redis
	go test -v -timeout 10s -race git.qutoutiao.net/gopher/qulibs/session
	go test -v -timeout 10s git.qutoutiao.net/gopher/qulibs/session
	go test -v -timeout 10s -race git.qutoutiao.net/gopher/qulibs/trace
	go test -v -timeout 10s git.qutoutiao.net/gopher/qulibs/trace

gorace: goclean
	go test -v -timeout 10s -race git.qutoutiao.net/gopher/qulibs/gorm
	go test -v -timeout 10s -race git.qutoutiao.net/gopher/qulibs/mongo
	go test -v -timeout 10s -race git.qutoutiao.net/gopher/qulibs/redis

gobench: goclean
	//go test -run=none -bench=.

govendor:
	mkdir -p ../qulibs-vendor
	rm -rf ./vendor && go mod vendor -v
	rm -rf ../qulibs-vendor/* && mv vendor/* ../qulibs-vendor/


