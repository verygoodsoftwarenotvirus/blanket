GOPATH     := $(GOPATH)
GIT_HASH   := $(shell git describe --tags --always --dirty)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%I:%M:%S%p')

.PHONY: coverage
coverage:
	if [ -f coverage.out ]; then rm coverage.out; fi
	go test -coverprofile=coverage.out
	go tool cover -html=coverage.out
	if [ -f coverage.out ]; then rm coverage.out; fi

.PHONY: blankoverage
blankoverage:
	if [ -f coverage.out ]; then rm coverage.out; fi
	go install
	go test -coverprofile=coverage.out
	blanket cover --html=coverage.out
	if [ -f coverage.out ]; then rm coverage.out; fi

.PHONY: introspect
introspect:
	go install
	blanket analyze --package=github.com/verygoodsoftwarenotvirus/blanket --fail-on-found

.PHONY: tests
tests:
	go test -v -cover -race

.PHONY: vendor
vendor:
	dep ensure -update -v

.PHONY: revendor
revendor:
	rm -rf vendor
	rm Gopkg.*
	dep init -v
