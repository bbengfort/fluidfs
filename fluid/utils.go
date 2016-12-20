// Provides utility and helper functions for the fluid package

package fluid

// Formatters for representing the date and time as a string.
const (
	JSONDateTime = "2006-01-02T15:04:05-07:00"
)

//===========================================================================
// Collection Helpers
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
