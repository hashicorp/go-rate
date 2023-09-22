.PHONY: all
all: test

.PHONY: tools
tools:
	go install github.com/hashicorp/copywrite@v0.15.0
	go install mvdan.cc/gofumpt@v0.3.1

.PHONY: test
test:
	go test -race -v ./...

.PHONY: cover-html
cover-html:
	go test -race -v -cover -coverprofile=.coverage ./... && \
		go tool cover -html=.coverage && \
		rm -f .coverage

.PHONY: bench
bench:
	go test -v -bench=. -count=1 -run=^#

.PHONY: copywrite
copywrite:
	copywrite headers

.PHONY: fmt
fmt:
	gofumpt -w .

.PHONY: gen
gen: copywrite fmt
