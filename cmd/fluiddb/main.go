// Database client to FluidFS metadata storage.
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bbengfort/fluidfs/fluid"
	"github.com/urfave/cli"

	fluiddb "github.com/bbengfort/fluidfs/fluid/db"
)

//===========================================================================
// Global Variables
//===========================================================================

var config *fluid.Config
var db fluiddb.Database
var buckets = []string{fluiddb.NamesBucket, fluiddb.PrefixesBucket, fluiddb.VersionsBucket}

//===========================================================================
// Main Method
//===========================================================================

func main() {
	// Load the .env file if it exists
	// godotenv.Load()

	// Instantiate the command line application
	app := cli.NewApp()
	app.Name = "fluiddb"
	app.Usage = "Database client to FluidFS metadata storage."
	app.Version = fluid.PackageVersion()

	// Global Flags
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "c, config",
			Usage: "specify a path to the yaml configuration",
		},
	}

	// Define the commands available to the app
	app.Commands = []cli.Command{
		{
			Name:   "delete",
			Usage:  "delete the configured database for a fresh start",
			Before: initConfig,
			Action: deleteDatabase,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "f, force",
					Usage: "don't ask before deleting",
				},
			},
		},
		{
			Name:      "count",
			Usage:     "count the number of items in the specified bucket(s)\n   if no buckets are specified, returns all bucket counts",
			ArgsUsage: "[bucket...]",
			Before:    initDatabase,
			Action:    countBucket,
		},
		{
			Name:      "get",
			Usage:     "print the value for the given key",
			ArgsUsage: "[key...]",
			Before:    initDatabase,
			Action:    viewKey,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "b, bucket",
					Usage: "specify the bucket the key is in",
					Value: fluiddb.NamesBucket,
				},
			},
		},
	}

	// Run the CLI program and parse the arguments
	app.Run(os.Args)
}

//===========================================================================
// Initialization Functions
//===========================================================================

// Initialize the configuration object
func initConfig(c *cli.Context) error {
	var err error

	// Load the configuration from YAML files on disk.
	config, err = fluid.LoadConfig(c.String("config"))
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	return nil
}

// Initialize a connection to the database.
func initDatabase(c *cli.Context) error {
	var err error

	if err = initConfig(c); err != nil {
		return err
	}

	// Initialize the database from the configuration.
	// Open a connection to the database
	db, err = fluiddb.InitDatabase(config.Database)
	if err != nil {
		err = fmt.Errorf("could not connect to database: %s", err.Error())
		return cli.NewExitError(err.Error(), 1)
	}

	// Log the connection to the database
	fmt.Printf("connected to %s\n", config.Database.String())
	return nil
}

//===========================================================================
// Command Handlers
//===========================================================================

// Delete the database for a fresh start.
func deleteDatabase(c *cli.Context) error {
	path := config.Database.Path

	if !c.Bool("force") {
		msg := fmt.Sprintf("Delete the database at %s?", path)
		confirm := promptConfirmation(msg)
		if !confirm {
			fmt.Println("Not deleting the database!")
			return nil
		}
	}

	err := os.Remove(path)
	if err != nil {
		msg := fmt.Sprintf("could not delete database at %s", path)
		return cli.NewExitError(msg, 1)
	}

	fmt.Printf("%s has been deleted\n", path)
	return nil
}

// Count the number of items in the specified buckets or all buckets
func countBucket(c *cli.Context) error {
	var buckets []string

	if c.NArg() == 0 {
		buckets = []string{fluiddb.NamesBucket, fluiddb.VersionsBucket, fluiddb.PrefixesBucket}
	} else {
		buckets = c.Args()
	}

	for _, bucket := range buckets {
		count, err := db.Count(bucket)
		if err != nil {
			return cli.NewExitError(err.Error(), 1)
		}

		fmt.Printf("%s: %d\n", bucket, count)
	}

	return nil
}

// Print out the JSON representation for the specified keys
func viewKey(c *cli.Context) error {
	bucket := c.String("bucket")
	if !fluid.ListContains(bucket, buckets) {
		msg := fmt.Sprintf("no bucket named '%s', use one of %s", bucket, strings.Join(buckets, ", "))
		return cli.NewExitError(msg, 1)
	}

	if c.NArg() < 1 {
		return cli.NewExitError("specify at least one key to get", 1)
	}

	for _, key := range c.Args() {
		val, err := db.Get([]byte(key), bucket)
		if err != nil {
			fmt.Printf("could not fetch key \"%s\": %s\n", key, err)
			continue
		}

		if val == nil {
			fmt.Printf("value not found for key \"%s\" in the %s bucket\n", key, bucket)
			continue
		}

		var pretty bytes.Buffer
		err = json.Indent(&pretty, val, "", "  ")
		if err != nil {
			fmt.Printf("could not indent json: %s\n", err)
			continue
		}

		fmt.Printf("Bucket: %s | Key: %s\n", bucket, key)
		fmt.Printf("%s\n", pretty.Bytes())
	}

	return nil
}

//===========================================================================
// Helper Functions
//===========================================================================

// Get confirmation from the command line to conduct an action.
func promptConfirmation(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [y/n]: ", prompt)

		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		response = strings.ToLower(strings.TrimSpace(response))

		switch response {
		case "yes":
			return true
		case "y":
			return true
		case "no":
			return false
		case "n":
			return false
		default:
			continue
		}
	}
}
