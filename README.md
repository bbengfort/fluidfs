# FluidFS

[![Build Status][travis_img]][travis_href]
[![Stories in Ready](https://img.shields.io/badge/waffle-ready-blue.svg)][waffle_href]
[![Documentation](https://img.shields.io/badge/docs-latest-green.svg)][docs]
![Version](https://img.shields.io/badge/version-alpha-red.svg)

[![Atlanta - Georgia Aquarium by Milos Kravcik][aquarium.jpg]][aquarium]

**A highly consistent distributed filesystem**

For more information, please see the documentation: [bbengfort.github.io/fluidfs][docs].

## Development

The primary interface is a command line program that interacts directly with the fluid library. Note that `cmd/fluid/main.go` uses the [CLI](https://github.com/urfave/cli) library rather than implementing console commands itself. Building from source is implemented using the included Makefile, which fetches dependencies and builds locally rather than to the `$GOPATH`:

    $ make

There is an RSpec-style test suite that uses [Ginkgo][ginkgo] and [Gomega](gomega). These tests can be run with the Makefile:

    $ make test

### Agile Board and Documentation

The development board can be found on Waffle:

- [https://waffle.io/bbengfort/fluidfs][waffle_href]

The documentation can be built and served locally with [mkdocs](http://www.mkdocs.org/):

    $ mkdocs serve

The latest version of the documentation is hosted with GitHub Pages and can be found at the project link: [bbengfort.github.io/fluidfs][docs]. To build and publish the documentation, use the make file:

    $ make publish

This will use the `mkdocs gh-deploy` command to build the site to the gh-pages branch and will push to origin.

## About

FluidFS is a research project to create a distributed file system in user space (FUSE) that is highly consistent and reliable. It is meant as a Dropbox replacement, allowing direct synchronization between devices on a personal network rather than going through a cloud service.

### Attribution

The image used in this README, ["Atlanta - Georgia Aquarium"][aquarium] by [Milos Kravcik](https://www.flickr.com/photos/49522551@N00/) is licensed under [CC-BY-NC-ND 2.0](https://creativecommons.org/licenses/by-nc-nd/2.0/).

<!-- Link References -->

[docs]: https://bbengfort.github.io/fluidfs/
[travis_img]: https://travis-ci.com/bbengfort/fluidfs.svg?token=5gAjQxGQg8bpYHKH9FmB
[travis_href]: https://travis-ci.com/bbengfort/fluidfs
[waffle_img]: https://badge.waffle.io/bbengfort/fluidfs.png?label=ready&title=Ready
[waffle_href]: https://waffle.io/bbengfort/fluidfs
[aquarium.jpg]: docs/img/aquarium.jpg
[aquarium]: https://flic.kr/p/aTUYyR
[ginkgo]: https://github.com/onsi/ginkgo
[gomega]: https://github.com/onsi/gomgea
