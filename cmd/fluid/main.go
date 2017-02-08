// CLI client to the FluidFS replica daemon.
package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/bbengfort/fluidfs/fluid"
	"github.com/urfave/cli"
)

// Client to make requests to the FluidFS server.
var client *fluid.CLIClient

// Handle the CLI application
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

	// Run before every command
	app.Before = initClient

	// Define the commands available to this helper.
	app.Commands = []cli.Command{
		{
			Name:     "status",
			Usage:    "check the status of the fluidfs server",
			Category: "client",
			Action:   fluidStatus,
		},
		{
			Name:     "mount",
			Usage:    "add a mount point to fluidfs",
			Category: "client",
			Action:   fluidMount,
		},
		{
			Name:     "web",
			Usage:    "get the url to the fluidfs web interface",
			Category: "client",
			Action:   fluidWeb,
		},
		{
			Name:      "chunk",
			Usage:     "debugging command to show chunk offsets",
			Category:  "debugging",
			ArgsUsage: "file [file ...]",
			Action:    fluidChunk,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "chunks, c",
					Usage: "specify fixed or variable chunking `method`",
				},
				cli.IntFlag{
					Name:  "blocksize, b",
					Usage: "specify the target block `size`",
				},
				cli.IntFlag{
					Name:  "minblocksize, l",
					Usage: "specify the minimum block `size`",
				},
				cli.IntFlag{
					Name:  "maxblocksize, u",
					Usage: "specify the maximum block `size`",
				},
			},
		},
	}

	// Run the CLI program and parse the arguments
	app.Run(os.Args)
}

//===========================================================================
// Helper Functions
//===========================================================================

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

//===========================================================================
// Client Commands
//===========================================================================

// Post a status request to the FluidFS server and write it out.
func fluidStatus(c *cli.Context) error {
	if err := client.Status(); err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	return nil
}

// Post a mount request to the FluidFS server.
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

// Post a request to get the address of the web interface and open a browser.
func fluidWeb(c *cli.Context) error {
	if err := client.Web(); err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	return nil
}

//===========================================================================
// Debugging Commands
//===========================================================================

// Use the chunker to show the offsets in the file specified.
func fluidChunk(c *cli.Context) error {
	if c.NArg() == 0 {
		return cli.NewExitError("specify files to chunk", 1)
	}

	// Create storage configuration
	conf := new(fluid.StorageConfig)
	conf.Defaults()

	if c.String("chunks") != "" {
		conf.Chunking = c.String("chunks")
	}

	if c.Int("blocksize") != 0 {
		conf.BlockSize = c.Int("blocksize")
	}

	if c.Int("minblocksize") != 0 {
		conf.BlockSize = c.Int("blocksize")
	}

	if c.Int("maxblocksize") != 0 {
		conf.BlockSize = c.Int("blocksize")
	}

	if err := conf.Validate(); err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	for idx, path := range c.Args() {
		// Read the file from the specified path
		data, err := ioutil.ReadFile(path)
		if err != nil {
			msg := fmt.Sprintf("could not read file at %s", path)
			return cli.NewExitError(msg, 1)
		}

		// Create the chunker using the FluidFS default configuration
		chunker, err := fluid.NewChunker(data, conf)
		if err != nil {
			return cli.NewExitError(err.Error(), 1)
		}

		// Begin output of chunker context
		fmt.Printf("File %d at %s\n", idx, path)

		blocks := 0
		offset := uint64(0)
		for chunker.Next() {
			blocks++
			stride := chunker.Offset()

			if conf.Chunking == fluid.VariableLengthChunking {
				fmt.Printf("   Block %d: %d to %d\n", blocks, offset, offset+stride)
			} else {
				fmt.Printf("   Block %d: %d to %d\n", blocks, offset, stride)
			}

			offset += stride
		}

	}

	return nil
}
