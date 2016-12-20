package fluid

import "fmt"

const (
	programName  = "fluidfs"
	majorVersion = 0
	minorVersion = 1
	microVersion = 0
	releaseLevel = "final"
)

// Version composes version information from the constants in this package
// and returns a string that defines current information about the package.
func Version() string {
	vstr := fmt.Sprintf("%d.%d", majorVersion, minorVersion)

	if microVersion > 0 {
		vstr += fmt.Sprintf(".%d", microVersion)
	}

	switch releaseLevel {
	case "final":
		return vstr
	case "alpha":
		return vstr + "a"
	case "beta":
		return vstr + "b"
	default:
		return vstr
	}

}
