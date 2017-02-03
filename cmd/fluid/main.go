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
	app.Version = fluid.PackageVersion()

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
		{
			Name:   "mount",
			Usage:  "add a mount point to fluidfs",
			Action: fluidMount,
		},
		{
			Name:   "web",
			Usage:  "get the url to the fluidfs web interface",
			Action: fluidWeb,
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

func fluidMount(c *cli.Context) error {
	if c.NArg() != 2 {
		return cli.NewExitError("mount requires mount point and prefix arguments", 1)
	}

	args := c.Args()
	if err := client.Mount(args[0], args[1]); err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	return nil
}

func fluidWeb(c *cli.Context) error {
	if err := client.Web(); err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	return nil
}
