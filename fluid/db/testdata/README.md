# Fluid Test Fixtures

This directory contains fixtures for testing FluidFS operation. The name and structure of the directory is described in [Test fixtures in Go](https://dave.cheney.net/2016/05/10/test-fixtures-in-go), which makes two primary points:

1. `go test` sets the working directory to the source directory of the package.
2. The Go tool will ignore any directory in $GOPATH named `testdata`.

Therefore opening a fixture file that is available on Travis and for local testing is as simple as follows:

```go
fixture := filepath.Join("testdata", "somefixture.json")
data, err := ioutil.ReadFile(fixture)
```

Note that this folder is for files and fixtures that should be committed to GitHub _before_ testing. Fixtures created _during_ testing should be placed in a temporary directory:

```go
tmpDir, err = ioutil.TempDir("", TempDirPrefix)
Ω(err).Should(BeNil())
```
