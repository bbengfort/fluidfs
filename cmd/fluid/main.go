// CLI client to the FluidFS replica daemon.
package main

import (
	"fmt"
	"os"

	"github.com/bbengfort/fluidfs/fluid"
	"github.com/urfave/cli"
)

var client *fluid.CLIClient

func main() {

	// Load the .env file if it exists
	// godotenv.Load()

	// Instantiate the command line application
	app := cli.NewApp()
	app.Name = "fluid"
	app.Usage = "FluidFS replica client for status and control."
	app.Version = fluid.Version()

	// Global flags
	app.Flags = []cli.Flag{}

	// Function run before every single command
	app.Before = initClient

	// Define the commands available to this helper.
	app.Commands = []cli.Command{
		{
			Name:   "status",
			Usage:  "check the status of the fluidfs server",
			Action: fluidStatus,
		},
	}

	// Run the CLI program and parse the arguments
	app.Run(os.Args)
}

func initClient(c *cli.Context) error {
	client = new(fluid.CLIClient)
	if err := client.Init(); err != nil {
		// If the FluidFS Server isn't running, just print the warning and
		// exit without an error code.
		fmt.Println(err.Error())
		os.Exit(0)
	}

	return nil
}

func fluidStatus(c *cli.Context) error {
	if err := client.Status(); err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	return nil
}
