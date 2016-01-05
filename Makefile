# Shell to use with Make
SHELL := /bin/bash

all: fmt deps
	@echo "Building FlowFS"
	@mkdir -p _bin/
	@go build -v -o _build/flow .

deps:
	@echo "Fetching dependencies"
	@go get -d -v ./flow/...

fmt:
	@echo "Formatting the source"
	gofmt -w .

test: deps
	ginkgo -r -v

citest: deps
	ginkgo -r -v --randomizeAllSpecs --randomizeSuites --failOnPending --cover --trace --race --compilers=2

clean:
	@echo "Cleaning up the project source."
	go clean

.PHONY:
	all deps fmt test citest
