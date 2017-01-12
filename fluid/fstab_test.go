package fluid_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	. "github.com/bbengfort/fluidfs/fluid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("fstab", func() {

	Describe("FSTable", func() {

		var err error
		var tmpDir string
		var fstab *FSTable

		BeforeEach(func() {
			tmpDir, err = ioutil.TempDir("", TempDirPrefix)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			fstab = new(FSTable)
		})

		AfterEach(func() {
			err = os.RemoveAll(tmpDir)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
		})

		It("should not return an error when loading an empty fstab", func() {
			path := filepath.Join(tmpDir, "fstab-does-not-exist")
			exists, _ := pathExists(path)
			Ω(exists).Should(BeFalse())

			err := fstab.Load(path)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
			Ω(fstab.Path).Should(Equal(path))
		})

		It("should initialize a default fstab on load", func() {
			Ω(fstab.Mounts).Should(BeZero())
			Ω(fstab.Path).Should(BeZero())
			Ω(fstab.Updated).Should(BeZero())

			err := fstab.Load("")
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			Ω(fstab.Mounts).ShouldNot(BeZero())
			Ω(fstab.Path).Should(BeZero())
			Ω(fstab.Updated).ShouldNot(BeZero())
		})

		// This will always pass, but use it to develop the fstabUpdateLine constant
		It("should be able to parse an update line", func() {
			line := "# FluidFS fstab config last updated: Wednesday, 11 Jan 2017 at 17:14:25 +0000\n"
			re, err := regexp.Compile(`^# FluidFS fstab config last updated: ([\w\d\s\-\+:,]+)$`)

			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
			Ω(re.MatchString(line)).Should(BeTrue())

			// Attempt to parse the date as well
			dtfmt := "Monday, 02 Jan 2006 at 15:04:05 -0700"
			sub := re.FindStringSubmatch(line)
			fmt.Println(sub)

			date, err := time.Parse(dtfmt, strings.TrimSpace(sub[1]))
			// time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
			expected := time.Date(2017, time.January, 11, 17, 14, 25, 0, time.UTC)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
			Ω(date.Equal(expected)).Should(BeTrue())
		})

		It("should be able to load a test fstab file", func() {
			path := filepath.Join("testdata", "fstab")
			fstab.Load(path)

			updated := time.Date(2017, time.January, 11, 22, 14, 25, 0, time.UTC)

			Ω(fstab.Path).Should(Equal(path))
			Ω(fstab.Mounts).Should(HaveLen(3))
			Ω(fstab.Updated.Equal(updated)).Should(BeTrue())

			// Check the number of comments on the mount points
			for i, mp := range fstab.Mounts {
				if i == 2 {
					Ω(mp.Comments).Should(HaveLen(3))
				} else {
					Ω(mp.Comments).Should(HaveLen(2))
				}
			}
		})

		It("should be able to save a test fstab file", func() {
			inpath := filepath.Join("testdata", "fstab")
			err := fstab.Load(inpath)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			outpath := filepath.Join(tmpDir, "fstab")
			err = fstab.Save(outpath)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
		})

	})

})
