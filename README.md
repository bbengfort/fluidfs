# FlowFS

[![Stories in Ready][waffle_img]][waffle_href]

**A highly consistent distributed filesystem built with FUSE**

[![Not with the flow by David Blackwell][planes.jpg]][planes]


## Development

The primary interface is a command line program that interacts directly with the flow library. Note that `main.go` uses the [CLI][cli] library rather than implementing console commands itself. Building from source is implemented using the included Makefile, which fetches dependencies and builds locally rather than to the `$GOPATH`.

There is an RSpec-style test suite that uses [Ginkgo][ginkgo] and [Gomega](gomega). These tests can be run with the Makefile by running `make test`.

## About

FlowFS is a research project to create a distributed file system in user space (FUSE) that is highly consistent and reliable. It is meant as a Dropbox replacement, allowing direct synchronization between devices on a personal network rather than going through a cloud service.

### Attribution

The image used in this README, ["Not with the flow"][planes] by [David Blackwell](https://www.flickr.com/photos/mobilestreetlife/) is licensed under [CC-BY-ND 2.0](https://creativecommons.org/licenses/by-nd/2.0/).

<!-- Link References -->

[travis_img]: https://travis-ci.org/bbengfort/flow.svg
[travis_href]: https://travis-ci.org/bbengfort/flow
[waffle_img]: https://badge.waffle.io/bbengfort/flow.png?label=ready&title=Ready
[waffle_href]: https://waffle.io/bbengfort/flow
[planes.jpg]: docs/img/planes.jpg
[planes]: https://flic.kr/p/gHrT81
[cli]: https://github.com/codegangsta/cli
[ginkgo]: https://github.com/onsi/ginkgo
[gomega]: https://github.com/onsi/gomgea
