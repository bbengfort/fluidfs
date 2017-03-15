# Shell to use with Make
SHELL := /bin/bash

# Export targets not associated with files.
.PHONY: all pkg deps fmt test citest clean publish doc protobuf

# Build FlowFS to a local build directory.
all: fmt deps
	@echo "Building FluidFS"
	@mkdir -p _bin/
	@go build -v -o _bin/fluid ./cmd/fluid
	@go build -v -o _bin/fluidfs ./cmd/fluidfs

# Build and package FlowFS to upload to GitHub
pkg: fmt
	@echo "Building and Packaging FluidFS"
	@mkdir -p fluidfs-darwin-amd64
	@go build -v -o fluidfs-darwin-amd64/fluid ./cmd/fluid
	@go build -v -o fluidfs-darwin-amd64/fluidfs ./cmd/fluidfs
	@cp fixtures/config-example.yml fluidfs-darwin-amd64/
	@zip -r fluidfs-darwin-amd64.zip fluidfs-darwin-amd64/
	@rm -rf fluidfs-darwin-amd64/

# Use godep to collect dependencies.
deps:
	@echo "Fetching dependencies"
	-godep restore

# Format the Go source code
fmt:
	@echo "Formatting the source"
	-gofmt -w .

# Target for simple testing on the command line
test:
	ginkgo -r -v

# Target for testing in continuous integration
citest:
	ginkgo -r -v --randomizeAllSpecs --randomizeSuites --failOnPending --cover --trace --race --compilers=2

# Run Godoc server and open browers
doc:
	- open http://localhost:6060/pkg/github.com/bbengfort/fluidfs/fluid/
	- godoc --http=:6060

# Clean build files
clean:
	@echo "Cleaning up the project source."
	-go clean
	-find . -name "*.coverprofile" -print0 | xargs -0 rm -rf
	-rm -rf site
	-rm -rf _bin
	-rm -rf _build

# Push documentation to GitHub Pages
publish:
	@echo "Deploying docs to gh-pages branch"
	@mkdocs gh-deploy --clean --quiet

# Compile protocol buffers
protobuf:
	@echo "Compiling protocol buffers"
	@protoc -I fluid/rpc/ fluid/rpc/*.proto --go_out=plugins=grpc:fluid/rpc
