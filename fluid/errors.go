package fluid

import "fmt"

// Error codes for various FluidFS error handling requirements
const (
	_                       = iota // Ignore Zero error codes
	ErrNotImplemented              // Feature is not currently implemented
	ErrImproperlyConfigured        // Configuration error or missing value
)

//===========================================================================
// Error Functions
//===========================================================================

// Errorf creates an error with the given code, prefix, and message but also
// performs some string formatting on behalf of the user (similar to the
// fmt.Errorf function, but with codes and prefixes).
func Errorf(message string, code int, prefix string, args ...interface{}) error {
	return &Error{
		Code:    code,
		Prefix:  prefix,
		Message: fmt.Sprintf(message, args...),
		err:     nil,
	}
}

// WrapError calls Errorf, but also includes the wrapped error in the return.
func WrapError(message string, code int, prefix string, err error, args ...interface{}) error {
	ferr := Errorf(message, code, prefix, args...).(*Error)
	ferr.err = err
	return ferr
}

// ImproperlyConfigured creates a new ErrImproperlyConfigured error.
func ImproperlyConfigured(message string, args ...interface{}) error {
	return Errorf(message, ErrImproperlyConfigured, "Improperly configured: ", args...)
}

//===========================================================================
// Error Type and Methods
//===========================================================================

// Error defines custom error handling for the fluid package.
type Error struct {
	Code    int    // The internal fluid error code
	Prefix  string // A prefix to append to the message
	Message string // The string description of the error
	err     error  // A wrapped error from another library
}

// Wraps returns true if the Error wraps another error.
func (err *Error) Wraps() bool {
	return err.err != nil
}

// Error implements the errors.Error interface.
func (err *Error) Error() string {
	if err.Wraps() {
		return fmt.Sprintf("%s%s: %s", err.Prefix, err.Message, err.err.Error())
	}
	return fmt.Sprintf("%s%s", err.Prefix, err.Message)
}
