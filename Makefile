GOPATH := $(shell go env GOPATH)
TMPDIR := $(shell mktemp -d)

all: checks

.PHONY: examples docs

checks: lint test examples functional-test

lint:
	@echo "Running $@ check"
	go tool golangci-lint run

vet: lint

test:
	@GO111MODULE=on SERVER_ENDPOINT=localhost:9000 ACCESS_KEY=obstoradmin SECRET_KEY=obstoradmin ENABLE_HTTPS=1 MINT_MODE=full go test -race -v ./...

examples:
	@echo "Building s3 examples"
	@cd ./examples/s3 && $(foreach v,$(wildcard examples/s3/*.go),go build -mod=mod -o ${TMPDIR}/$(basename $(v)) $(notdir $(v)) || exit 1;)
	@echo "Building obstor examples"
	@cd ./examples/obstor && $(foreach v,$(wildcard examples/obstor/*.go),go build -mod=mod -o ${TMPDIR}/$(basename $(v)) $(notdir $(v)) || exit 1;)

functional-test:
	@GO111MODULE=on go build -race functional_tests.go
	@SERVER_ENDPOINT=localhost:9000 ACCESS_KEY=obstoradmin SECRET_KEY=obstoradmin ENABLE_HTTPS=1 MINT_MODE=full ./functional_tests

functional-test-notls:
	@GO111MODULE=on go build -race functional_tests.go
	@SERVER_ENDPOINT=localhost:9000 ACCESS_KEY=obstoradmin SECRET_KEY=obstoradmin ENABLE_HTTPS=0 MINT_MODE=full ./functional_tests

clean:
	@echo "Cleaning up all the generated files"
	@find . -name '*.test' | xargs rm -fv
	@find . -name '*~' | xargs rm -fv
