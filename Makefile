GOPATH            := $(GOPATH)
GIT_HASH          := $(shell git describe --tags --always --dirty)
BUILD_TIME        := $(shell date -u '+%Y-%m-%d_%I:%M:%S%p')
TESTABLE_PACKAGES := $(shell go list gitlab.com/verygoodsoftwarenotvirus/blanket/... | grep -v -e "example_packages")

.PHONY: binary
binary:
	go build -o devBlanket gitlab.com/verygoodsoftwarenotvirus/blanket/cmd/blanket

.PHONY: blankoverage
blankoverage: binary
	if [ -f coverage.out ]; then rm coverage.out; fi
		go test -coverprofile=coverage.out
		blanket cover --html=coverage.out
	if [ -f coverage.out ]; then rm coverage.out; fi

.PHONY: introspect
introspect: binary
	for pkg in $(TESTABLE_PACKAGES); do \
		set -e; \
		./devBlanket analyze --fail-on-found --package=$$pkg; \
	done

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
	set -ex; go test -v -cover -race $(TESTABLE_PACKAGES)

.PHONY: coverage
coverage:
	if [ -f coverage.out ]; then rm coverage.out; fi
	echo "mode: set" > coverage.out

	for pkg in $(TESTABLE_PACKAGES); do \
		set -e; \
		go test -coverprofile=profile.out -v -race $$pkg; \
		cat profile.out | grep -v "mode: atomic" >> coverage.out; \
	done
	rm profile.out

.PHONY: ci-coverage
ci-coverage:
	go test $(TESTABLE_PACKAGES) -v -coverprofile=profile.out

.PHONY: docker-image
docker-image:
	docker build --tag 'blanket:latest' .
