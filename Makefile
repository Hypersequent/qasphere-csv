GOCMD=GO111MODULE=on go

linters-install:
	@golangci-lint --version >/dev/null 2>&1 || { \
		echo "installing linting tools..."; \
		curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v2.12.2; \
	}

lint: linters-install
	golangci-lint run

test:
	$(GOCMD) test -v -cover -race $$(go list ./... | grep -v /examples/)

build-examples:
	$(GOCMD) build ./examples/...

.PHONY: test lint linters-install build-examples
