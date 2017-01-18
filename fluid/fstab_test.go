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
	"github.com/google/uuid"

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

	Describe("MountPoint", func() {

		It("should be able to parse mount point definitions", func() {

			parseTests := []string{
				"8859b5c7-d860-11e6-9b0a-28cfe91c6851 /data/mnt/bravo bar 501 22 red,auto,blot 0 1",
				"8859b5c7-d860-11e6-9b0a-28cfe91c6851\t/data/mnt/bravo\tbar\t501\t22\tred,auto,blot\t0\t1",
				" 8859b5c7-d860-11e6-9b0a-28cfe91c6851 /data/mnt/bravo bar 501 22 red,auto,blot 0 1",
				"\t8859b5c7-d860-11e6-9b0a-28cfe91c6851\t/data/mnt/bravo\tbar\t501\t22\tred,auto,blot\t0\t1\t",
				"8859b5c7-d860-11e6-9b0a-28cfe91c6851    /data/mnt/bravo  bar 501     22  red,auto,blot     0     1  ",
				"8859b5c7-d860-11e6-9b0a-28cfe91c6851 \t /data/mnt/bravo \t bar \t 501 \t 22 \t red,auto,blot \t 0 \t 1",
			}

			guid, _ := uuid.Parse("8859b5c7-d860-11e6-9b0a-28cfe91c6851")
			mp := &MountPoint{
				guid, "/data/mnt/bravo", "bar",
				501, 22, false, true,
				nil, []string{"red", "auto", "blot"},
			}

			for _, pt := range parseTests {
				tmp := new(MountPoint)
				err := tmp.Parse(pt)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(tmp).Should(Equal(mp))
			}

		})

		It("should not parse lines with an incorrect number of fields", func() {
			badLines := []string{
				"8859b5c7-d860-11e6-9b0a-28cfe91c6851 /data/mnt/bravo bar 501 22 default 0 1 1 0",
				"/data/mnt/bravo bar 501 22 default 0 1",
				"8859b5c7-d860-11e6-9b0a-28cfe91c6851 /data/mnt/bravo bar baz 501 22 default 0 1",
			}

			for _, bl := range badLines {
				mp := new(MountPoint)
				err := mp.Parse(bl)
				Ω(err).Should(MatchError("could not parse mount point: not enough fields"))
			}
		})

		It("should not parse lines with a malformed UUID", func() {
			badLines := []string{
				"/data/mnt/bravo bar 501 22 default 0 1 6",
				"8859b5c7 /data/mnt/bravo bar 501 22 default 0 1",
			}

			for _, bl := range badLines {
				mp := new(MountPoint)
				err := mp.Parse(bl)
				Ω(err).Should(MatchError(MatchRegexp(`could not parse UUID field: invalid UUID length: \d+`)))
			}
		})

		It("should not parse lines with non-int UID and GID fields", func() {
			badLines := []string{
				"8859b5c7-d860-11e6-9b0a-28cfe91c6851 /data/mnt/bravo bar ubuntu 22 red,auto,blot 0 1",
				"8859b5c7-d860-11e6-9b0a-28cfe91c6851 /data/mnt/bravo bar 501 ubuntu red,auto,blot 0 1",
			}

			for _, bl := range badLines {
				mp := new(MountPoint)
				err := mp.Parse(bl)
				Ω(err).Should(MatchError(MatchRegexp(`could not parse [UG]ID field:`)))
			}
		})

		It("should parse comma separated options", func() {
			var lines = []struct {
				line string
				opts []string
			}{
				{"8859b5c7-d860-11e6-9b0a-28cfe91c6851 /data/mnt/bravo bar 501 22 red,auto,blot 0 1", []string{"red", "auto", "blot"}},
				{"8859b5c7-d860-11e6-9b0a-28cfe91c6851 /data/mnt/bravo bar 501 22 defaults 0 1", []string{"defaults"}},
			}

			for _, tt := range lines {
				mp := new(MountPoint)
				err := mp.Parse(tt.line)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(mp.Options).Should(Equal(tt.opts))
			}

		})

		It("should be able to parse booleans", func() {
			var lines = []struct {
				line  string
				bools []bool
			}{
				{"8859b5c7-d860-11e6-9b0a-28cfe91c6851 /data/mnt/bravo bar 501 22 red,auto,blot 0 1", []bool{false, true}},
				{"8859b5c7-d860-11e6-9b0a-28cfe91c6851 /data/mnt/bravo bar 501 22 red,auto,blot 1 1", []bool{true, true}},
				{"8859b5c7-d860-11e6-9b0a-28cfe91c6851 /data/mnt/bravo bar 501 22 red,auto,blot 0 0", []bool{false, false}},
				{"8859b5c7-d860-11e6-9b0a-28cfe91c6851 /data/mnt/bravo bar 501 22 red,auto,blot 1 0", []bool{true, false}},
				{"8859b5c7-d860-11e6-9b0a-28cfe91c6851 /data/mnt/bravo bar 501 22 red,auto,blot t f", []bool{true, false}},
				{"8859b5c7-d860-11e6-9b0a-28cfe91c6851 /data/mnt/bravo bar 501 22 red,auto,blot T F", []bool{true, false}},
				{"8859b5c7-d860-11e6-9b0a-28cfe91c6851 /data/mnt/bravo bar 501 22 red,auto,blot true false", []bool{true, false}},
			}

			for _, tt := range lines {
				mp := new(MountPoint)
				err := mp.Parse(tt.line)
				Ω(err).ShouldNot(HaveOccurred())
				Ω([]bool{mp.Store, mp.Replicate}).Should(Equal(tt.bools))
			}

		})

		It("should raise an exception when bools cannot be parsed", func() {
			badLines := []string{
				"8859b5c7-d860-11e6-9b0a-28cfe91c6851 /data/mnt/bravo bar 501 22 defaults a 1",
				"8859b5c7-d860-11e6-9b0a-28cfe91c6851 /data/mnt/bravo bar 501 22 defaults 0 b",
			}

			for _, bl := range badLines {
				mp := new(MountPoint)
				err := mp.Parse(bl)
				Ω(err).Should(MatchError(MatchRegexp(`could not parse Store|Replicate field:`)))
			}
		})

		It("should be able to stringify a mount point", func() {
			mps := "8859b5c7-d860-11e6-9b0a-28cfe91c6851 /data/mnt/bravo bar 501 22 defaults false true"
			mp := new(MountPoint)
			err := mp.Parse(mps)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(mp.String()).Should(Equal(mps))
		})

	})

})
