// Handling for Lamport Scalar and other version numbers.

package fluid

import "fmt"

//===========================================================================
// Lamport Scalar Version Type
//===========================================================================

// NewVersion creates a new version sequence with the local precedence ID.
func NewVersion() *Version {
	return &Version{config.PID, 0, 0}
}

// Version implements a Lamport scalar version number that has two components:
// the process id and the scalar, a montonically increasing counter. Versions
// can be updated from another version object, which will increase the scalar
// to the maximal value of the two scalars, minimizing cross-process conflict.
//
// Versions are compared from right to left. E.g. if a version is represented
// as (pid, scalar), then the scalars are compared first, if they are equal,
// then the pids are compared.
type Version struct {
	PID    uint   // Process or Precendence ID (assigned per replica)
	Scalar uint64 // Montononically increasing scalar value
	Latest uint64 // The latest observed value, used to calculate next.
}

// IsRoot returns true if the version has a zero value for the scalar, meaning
// it is a version that was created but cannot be assigned to a file.
func (v *Version) IsRoot() bool {
	return v.Scalar == 0
}

// Update the version with the latest seen, computed as the maximal scalar
// between the two versions. Note that this does reconcile and modify both
// version' latest value.
func (v *Version) Update(o *Version) {
	latest := MaxUInt64(v.Latest, o.Latest)
	v.Latest = latest
	o.Latest = latest
}

// Next returns a new Version with the scaler = 1 + latest. If a zero value
// PID is passed in, then the PID from the previous version is maintained.
// This method also updates the latest version of the previous version.
func (v *Version) Next(pid uint) *Version {
	// Mantain PID if needed.
	if pid == 0 {
		pid = v.PID
	}

	// Increment the latest version (on the old version)
	v.Latest++

	// Create a new Version from the latest and return it
	return &Version{pid, v.Latest, v.Latest}
}

// String representation of a version
func (v *Version) String() string {
	return fmt.Sprintf("(%d, %d)", v.Scalar, v.PID)
}

//===========================================================================
// Lamport Scalar Version Comparision
//===========================================================================

// LessEqual returns true if the local version is lass than or equal to
// the other version.
func (v *Version) LessEqual(o *Version) bool {
	if v.Scalar == o.Scalar {
		return v.PID <= o.PID
	}
	return v.Scalar <= o.Scalar
}

// Less returns true if the local version is less than the other version.
func (v *Version) Less(o *Version) bool {
	if v.Scalar == o.Scalar {
		return v.PID < o.PID
	}
	return v.Scalar < o.Scalar
}

// Equal returns true if the versions have the same scalar and pid.
func (v *Version) Equal(o *Version) bool {
	return v.PID == o.PID && v.Scalar == o.Scalar
}

// NotEqual returns true if the versions do not have the same scalar and pid.
func (v *Version) NotEqual(o *Version) bool {
	return !v.Equal(o)
}

// Greater returns true if the local version is greater than the other version.
func (v *Version) Greater(o *Version) bool {
	if v.Scalar == o.Scalar {
		return v.PID > o.PID
	}
	return v.Scalar > o.Scalar
}

// GreaterEqual returns true if the local version is greater than or equal to
// the other version.
func (v *Version) GreaterEqual(o *Version) bool {
	if v.Scalar == o.Scalar {
		return v.PID >= o.PID
	}
	return v.Scalar >= o.Scalar
}
