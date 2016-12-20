// CLI client to the FluidFS replica daemon.
package main

import (
	"os"

	"github.com/bbengfort/fluidfs/fluid"
	"github.com/urfave/cli"
)

func main() {

	app := cli.NewApp()
	app.Name = "fluid"
	app.Usage = "FluidFS replica client for status and control."
	app.Version = fluid.Version()
	app.Author = "Benjamin Bengfort"
	app.Email = "bengfort@cs.umd.edu"

	app.Run(os.Args)
}
