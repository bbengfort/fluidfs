package fluid_test

import (
	"math/rand"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

const TempDirPrefix = "com.fluidfs."

func TestFluid(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fluid Suite")
}

//===========================================================================
// Testing Helper Functions
//===========================================================================

// In-place reverse a list of byte slices.
func reverse(list [][]byte) {
	for i := len(list)/2 - 1; i >= 0; i-- {
		j := len(list) - 1 - i
		list[i], list[j] = list[j], list[i]
	}
}

// Reverse a list of strings using the sort package
func sreverse(list []string) {
	for i := len(list)/2 - 1; i >= 0; i-- {
		j := len(list) - 1 - i
		list[i], list[j] = list[j], list[i]
	}
}

// Runes for the random string function
var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// Create a random string of length n
func randString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// Check if a path exists
func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return true, nil
}
