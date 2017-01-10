package fluid_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"

	. "github.com/bbengfort/fluidfs/fluid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// This function skips if a PID file exists. There are a whole suite of tests
// that require the PID file including the CLI tests, etc.
func SkipIfPIDExists() {
	pid := new(PID)
	if exists, _ := pathExists(pid.Path()); exists {
		Skip("cannot run CLI tests if a FluidFS server is already running")
	}
}

var _ = Describe("PID File", func() {

	var pid *PID

	BeforeEach(func() {
		pid = new(PID)
	})

	It("should return a path to the PID file", func() {
		Ω(pid.Path()).ShouldNot(BeZero())
	})

	It("should be able to select a free and open port", func() {
		port, err := pid.FreePort()

		Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
		Ω(port).ShouldNot(BeZero())

		// Create a listener on the port
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		Ω(err).Should(BeNil(), fmt.Sprintf("can't listen on port %d: %s", port, err))

		// Close the listener
		Ω(ln.Close()).Should(BeNil(), fmt.Sprintf("can't stop listening on port: %d: %s", port, err))
	})

	It("should be able to compose an address with the port", func() {
		pid.PID = 42
		pid.PPID = 41
		pid.Port = 6060

		Ω(pid.Addr()).Should(Equal("localhost:6060"))
	})

	Context("tests only if the PID file doesn't exist", func() {

		freePID := false

		BeforeEach(func() {
			// If the PID file does not exist before tests, then remove it after.
			if exists, _ := pathExists(pid.Path()); !exists {
				freePID = true
			}
		})

		AfterEach(func() {
			if freePID {
				// Remove the PID file if it doesn't exist
				os.Remove(pid.Path())
			}
		})

		It("should be able to save the PID file with selected info", func() {
			SkipIfPIDExists()

			// Ensure the PID file does not exist
			exists, _ := pathExists(pid.Path())
			Ω(exists).Should(BeFalse())

			// Save the PID file
			err := pid.Save()
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			// Ensure that the PID is populated
			Ω(pid.PID).ShouldNot(BeZero())
			Ω(pid.PPID).ShouldNot(BeZero())
			Ω(pid.Port).ShouldNot(BeZero())

			// Ensure that the PID file exists
			exists, _ = pathExists(pid.Path())
			Ω(exists).Should(BeTrue())
		})

		It("should be able to load the PID file", func() {
			SkipIfPIDExists()

			// Write a test PID file
			testData := map[string]int{
				"pid":  23,
				"ppid": 22,
				"port": 50800,
			}

			// Write the test data as JSON
			data, err := json.Marshal(testData)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			// And spit it out to a file.
			os.MkdirAll(filepath.Dir(pid.Path()), ModeStorageDir)
			err = ioutil.WriteFile(pid.Path(), data, ModeBlob)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			// Make sure the PID is zeroed out.
			Ω(pid.PID).Should(BeZero())
			Ω(pid.PPID).Should(BeZero())
			Ω(pid.Port).Should(BeZero())

			// Load the PID file
			err = pid.Load()
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			// Ensure that the PID is loaded
			Ω(pid.PID).Should(Equal(23))
			Ω(pid.PPID).Should(Equal(22))
			Ω(pid.Port).Should(Equal(50800))
		})

		It("should be able to delete the PID file", func() {
			SkipIfPIDExists()

			exists, _ := pathExists(pid.Path())
			Ω(exists).Should(BeFalse())

			pid.Save()

			exists, _ = pathExists(pid.Path())
			Ω(exists).Should(BeTrue())

			pid.Free()

			exists, _ = pathExists(pid.Path())
			Ω(exists).Should(BeFalse())
		})

	})

})
