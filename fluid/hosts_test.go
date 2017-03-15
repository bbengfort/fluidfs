package fluid_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	. "github.com/bbengfort/fluidfs/fluid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("hosts", func() {

	Describe("Hosts", func() {

		var err error
		var tmpDir string
		var hosts *Hosts

		BeforeEach(func() {
			tmpDir, err = ioutil.TempDir("", TempDirPrefix)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			hosts = new(Hosts)
		})

		AfterEach(func() {
			err = os.RemoveAll(tmpDir)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
		})

		It("should not return an error when loading an empty hosts file", func() {
			path := filepath.Join(tmpDir, "hosts-does-not-exist")
			exists, _ := pathExists(path)
			Ω(exists).Should(BeFalse())

			err := hosts.Load(path)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
			Ω(hosts.Path).Should(Equal(path))
		})

		It("should initialize default hosts on load", func() {
			Ω(hosts.Replicas).Should(BeZero())
			Ω(hosts.Path).Should(BeZero())
			Ω(hosts.Updated).Should(BeZero())

			err := hosts.Load("")
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			Ω(hosts.Replicas).ShouldNot(BeZero())
			Ω(hosts.Path).Should(BeZero())
			Ω(hosts.Updated).ShouldNot(BeZero())
		})

		It("should be able to load a test hosts file", func() {
			path := filepath.Join("testdata", "hosts")
			err := hosts.Load(path)
			msg := fmt.Sprintf("could not load hosts file: %s", err)
			Ω(err).ShouldNot(HaveOccurred(), msg)

			Ω(hosts.Path).Should(Equal(path))
			Ω(hosts.Replicas).Should(HaveLen(5))
		})

		It("should be able to save a test hosts file", func() {
			// Create some test data
			hosts := &Hosts{
				Path:    "",
				Updated: time.Now(),
				Replicas: map[string]*Replica{
					"apollo": &Replica{
						1, "apollo", "192.168.1.12", 4157, false, 23, 2, make([]string, 0), time.Now(), time.Now(), time.Time{}, 0, 0, "",
					},
					"cyrus": &Replica{
						2, "cyrus", "192.168.1.13", 4157, false, 23, 2, make([]string, 0), time.Now(), time.Now(), time.Time{}, 0, 0, "",
					},
					"sol": &Replica{
						3, "sol", "192.168.1.14", 4157, false, 23, 2, make([]string, 0), time.Now(), time.Now(), time.Time{}, 0, 0, "",
					},
				},
			}

			// Save the hosts
			outpath := filepath.Join(tmpDir, "hosts")
			err = hosts.Save(outpath)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			// Make sure the path exists
			exists, err := pathExists(outpath)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
			Ω(exists).Should(BeTrue(), "no hosts at designated outpath")

			// Load the data from the saved hosts
			hosts2 := new(Hosts)
			err = hosts2.Load(outpath)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			// New hosts should be indentical to the old hosts
			for name, replica := range hosts.Replicas {
				Ω(hosts2.Replicas).Should(HaveKey(name))
				Ω(hosts2.Replicas[name]).Should(Equal(replica))
			}

			// There should be no additional hosts
			Ω(hosts2.Replicas).Should(HaveLen(len(hosts.Replicas)))

		})

	})

})
