package db_test

import (
	"bufio"
	"errors"
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/bbengfort/fluidfs/fluid/db"

	"testing"
)

const TempDirPrefix = "com.fluidfs.db."

func TestDb(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Database Suite")
}

// Helper function to load a fixture to the database
func loadDBFixture(db Database, bucket string, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), "\t")
		if len(parts) != 2 {
			return errors.New("could not parse key/value line")
		}

		db.Put([]byte(parts[0]), []byte(parts[1]), bucket)
	}

	if err = scanner.Err(); err != nil {
		return err
	}

	return nil
}
