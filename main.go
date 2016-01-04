// Command line interface to FlowFS

package main

import (
    "os"

    // "github.com/bbengfort/flow/flow"
    "github.com/bbengfort/flow/flow/version"
    "github.com/codegangsta/cli"
)

func main() {

    app := cli.NewApp()
    app.Name  = "flow"
    app.Usage = "A highly consistent distributed filesystem built with FUSE."
    app.Version = version.Version()
    app.Author = "Benjamin Bengfort"
    app.Email  = "bengfort@cs.umd.edu"

    app.Run(os.Args)
}
