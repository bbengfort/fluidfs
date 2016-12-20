// FluidFS replica daemon process
package main

import (
	"os"

	"github.com/bbengfort/fluidfs/fluid"
	"github.com/urfave/cli"
)

func main() {

	app := cli.NewApp()
	app.Name = "fluidfs"
	app.Usage = "A highly consistent distributed filesystem."
	app.Version = fluid.Version()
	app.Author = "Benjamin Bengfort"
	app.Email = "bengfort@cs.umd.edu"

	app.Run(os.Args)
}
