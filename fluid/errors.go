package fluid

import "fmt"

// Error defines custom error handling for the fluid package.
type Error struct {
	Code    int    // The internal fluid error code
	Message string // The string description of the error
}

// Error implements the errors.Error interface.
func (err *Error) Error() string {
	return fmt.Sprintf("Error %d: %s", err.Code, err.Message)
}
