// Provides utility and helper functions for the fluid package

package fluid

import (
	"os"
	"strings"
)

// Formatters for representing the date and time as a string.
const (
	JSONDateTime = "2006-01-02T15:04:05-07:00"
)

//===========================================================================
// String Helpers
//===========================================================================

// Regularize a string for comparision, e.g. make all lowercase and trim
// whitespace. This utility is often used on user input to compare to constant
// strings like database drivers or hashing algorithm names.
func Regularize(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ToLower(value)
	return value
}

// Stride returns an array of N length substrings.
func Stride(s string, n int) []string {

	a := []rune(s) // Convert string to a slice of runes

	// Compute the length of the output array
	l := len(a) / n
	if len(a)%n != 0 {
		l++
	}

	o := make([]string, 0, l) // Create the output array

	// Range over the runes by n strides and append strings to output.
	for i := 0; i < len(a); i = i + n {
		j := i + n
		if j > len(a) {
			j = len(a)
		}
		o = append(o, string(a[i:j]))
	}

	return o
}

// StrideFixed returns an array of N length substrings and does not allow the
// last element to have a length < N (e.g. no remainders).
func StrideFixed(s string, n int) []string {
	o := Stride(s, n)

	// If the length of the last element is less than n, don't return it
	if len(o[len(o)-1]) < n {
		return o[:len(o)-1]
	}

	return o
}

//===========================================================================
// String Collection Helpers
//===========================================================================

// ListContains searches a list for a particular value in O(n) time.
func ListContains(value string, list []string) bool {
	for _, elem := range list {
		if elem == value {
			return true
		}
	}
	return false
}

// IndexOf searches a list for a particular value in O(n) time.
func IndexOf(value string, list []string) int {
	for pos, val := range list {
		if val == value {
			return pos
		}
	}
	return -1
}

// Remove an element from a string collection
func Remove(value string, list []string) []string {
	idx := IndexOf(value, list)
	if idx < 0 {
		return list
	}

	return append(list[:idx], list[idx+1:]...)
}

//===========================================================================
// Numeric Helpers
//===========================================================================

// MaxUInt64 returns the maximal value of the list of passed in uints
func MaxUInt64(values ...uint64) uint64 {
	max := uint64(0) // this works because values are unsigned.
	for _, val := range values {
		if val > max {
			max = val
		}
	}
	return max
}

// Blocks returns the number of 512 byte blocks required
func Blocks(value uint64) uint64 {
	if value == 0 {
		return 0
	}

	blocks := value / minBlockSize
	if value%minBlockSize > 0 {
		return blocks + 1
	}

	return blocks
}

//===========================================================================
// File Path Helpers
//===========================================================================

// IsHidden returns true if the file name starts with "." or "~" or if a
// directory name starts with "." (skip "~" in case it's the home directory).
// This method does not work on Windows currently.
func IsHidden(info os.FileInfo) bool {
	name := info.Name()

	if info.IsDir() {
		if strings.HasPrefix(name, ".") {
			return true
		}

		return false
	}

	if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "~") {
		return true
	}

	return false
}
