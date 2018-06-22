GOPATH     := $(GOPATH)
GIT_HASH   := $(shell git describe --tags --always --dirty)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%I:%M:%S%p')
TESTABLE_PACKAGES = $(shell go list github.com/verygoodsoftwarenotvirus/blanket/... | grep -v -e "example_packages")

clean:
	rm -f blanket

.PHONY: blanket
blanket: clean
	go build -o blanket github.com/verygoodsoftwarenotvirus/blanket/cmd/blanket

.PHONY: blankoverage
blankoverage: blanket
	if [ -f coverage.out ]; then rm coverage.out; fi
		go test -coverprofile=coverage.out
		blanket cover --html=coverage.out
	if [ -f coverage.out ]; then rm coverage.out; fi

.PHONY: introspect
introspect: blanket
	# for pkg in $(TESTABLE_PACKAGES); do \
	# 	set -e; \
	# 	./blanket analyze --fail-on-found --package= $$pkg; \
	# done

	./blanket analyze --fail-on-found --package=github.com/verygoodsoftwarenotvirus/blanket/cmd/blanket
	./blanket analyze --fail-on-found --package=github.com/verygoodsoftwarenotvirus/blanket/lib/util
	./blanket analyze --fail-on-found --package=github.com/verygoodsoftwarenotvirus/blanket/output/html
	# ./blanket analyze --fail-on-found --package=github.com/verygoodsoftwarenotvirus/blanket/analysis

.PHONY: vendor
vendor:
	dep ensure -update -v

.PHONY: revendor
revendor:
	rm -rf vendor
	rm Gopkg.*
	dep init -v

.PHONY: tests
tests:
	set -ex; go test -v -cover -race $(shell go list github.com/verygoodsoftwarenotvirus/blanket/... | grep -v -e "example_packages")

.PHONY: coverage
coverage:
	if [ -f coverage.out ]; then rm coverage.out; fi
	echo "mode: set" > coverage.out

	for pkg in $(TESTABLE_PACKAGES); do \
		set -e; \
		go test -coverprofile=profile.out -v -cover -race $$pkg; \
		cat profile.out | grep -v "mode: set" >> coverage.out; \
	done
	rm profile.out

.PHONY: docker-image
docker-image:
	docker build --tag 'blanket:latest' .
