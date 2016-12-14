// Command line interface to FluidFS

package main

import (
    "os"

    "github.com/bbengfort/fluidfs/fluid/version"
    "github.com/codegangsta/cli"
)

func main() {

    app := cli.NewApp()
    app.Name  = "fluid"
    app.Usage = "A highly consistent distributed filesystem built with FUSE."
    app.Version = version.Version()
    app.Author = "Benjamin Bengfort"
    app.Email  = "bengfort@cs.umd.edu"

    app.Run(os.Args)
}
