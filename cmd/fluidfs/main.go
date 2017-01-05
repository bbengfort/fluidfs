// FluidFS replica daemon process
package main

import (
	"fmt"
	"os"

	"github.com/bbengfort/fluidfs/fluid"
	"github.com/urfave/cli"
)

var server *fluid.Server

func main() {

	// Load the .env file if it exists
	// godotenv.Load()

	// Instantiate the command line application
	app := cli.NewApp()
	app.Name = "fluidfs"
	app.Usage = "A highly consistent distributed filesystem."
	app.Version = fluid.Version()

	// Global flags
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "c, config",
			Usage: "specify the path to a yaml configuration",
		},
	}

	// Function run before every single command
	app.Before = initFluid

	// Define the commands available to this helper.
	app.Commands = []cli.Command{
		{
			Name:   "start",
			Usage:  "start the fluidfs server",
			Action: startServer,
		},
		{
			Name:   "config",
			Usage:  "print the configuration and exit",
			Action: printConfig,
		},
	}

	// Run the CLI program and parse the arguments
	app.Run(os.Args)
}

func initFluid(c *cli.Context) error {
	server = new(fluid.Server)
	if err := server.Init(c.String("config")); err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	return nil
}

func startServer(c *cli.Context) error {
	if err := server.Run(); err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	return nil
}

func printConfig(c *cli.Context) {
	// Print the configuration and exit
	fmt.Println(server.Config.String())
}
