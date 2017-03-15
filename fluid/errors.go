package fluid

import "fmt"

// Error codes for various FluidFS error handling requirements
const (
	_                       = iota // Ignore Zero error codes
	ErrFluidExit                   // FluidFS must exit (default error)
	ErrFileSystem                  // Error with FUSE or the file system
	ErrNotImplemented              // Feature is not currently implemented
	ErrImproperlyConfigured        // Configuration error or missing value
	ErrNetworkUnavailable          // Cannot reach network interface
	ErrUninitialized               // A required component is not initialized correctly
	ErrUnsupported                 // The operating system or component is not supported
	ErrChunking                    // Something went wrong during chunking
	ErrAPIAccess                   // Soemthing went wrong accessing the API
	ErrParsing                     // Something went wrong parsing a file
	ErrInvalidReplica              // A replica is not configured correctly
	ErrDatabase                    // Error interacting with the database
	ErrBlobStorage                 // Error with blob storage
	ErrReplication                 // Error with blob or version replication
)

//===========================================================================
// Error Functions
//===========================================================================

// NewError creates a new simple error with the given code and prefix.
func NewError(message string, code int, prefix string) error {
	// Set the default error code
	if code == 0 {
		code = ErrFluidExit
	}

	// Create the error and return
	return &Error{
		Code:    code,
		Prefix:  prefix,
		Message: message,
		err:     nil,
	}
}

// Errorc creates a simple error with the given code and no prefix.
func Errorc(message string, code int) error {
	return NewError(message, code, "")
}

// Errors creates a simple error with the default code and no prefix.
func Errors(message string) error {
	return Errorc(message, ErrFluidExit)
}

// Errorsf creates a simple error with string formatting
func Errorsf(message string, args ...interface{}) error {
	return Errors(fmt.Sprintf(message, args...))
}

// Errorw creates a simple wrapped error with default code and no prefix.
func Errorw(message string, err error) error {
	return WrapError(message, ErrFluidExit, "", err)
}

// Errorf creates an error with the given code, prefix, and message but also
// performs some string formatting on behalf of the user (similar to the
// fmt.Errorf function, but with codes and prefixes).
func Errorf(message string, code int, prefix string, args ...interface{}) error {
	return NewError(fmt.Sprintf(message, args...), code, prefix)
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

// NetworkError creates a new ErrNetworkUnavailable error.
func NetworkError(message string, err error, args ...interface{}) error {
	return WrapError(message, ErrNetworkUnavailable, "Network unavailable: ", err, args...)
}

// ChunkingError creates a new ErrChunking error.
func ChunkingError(message string, args ...interface{}) error {
	return Errorf(message, ErrChunking, "", args...)
}

// APIAccessError creates a new ErrAPIAccess error.
func APIAccessError(message string, err error, args ...interface{}) error {
	return WrapError(message, ErrAPIAccess, "", err, args...)
}

// ParsingError creates a new ErrParsing error.
func ParsingError(message string, err error, args ...interface{}) error {
	return WrapError(message, ErrParsing, "", err, args...)
}

// DatabaseError creates a new ErrDatabase error.
func DatabaseError(message string, err error, args ...interface{}) error {
	return WrapError(message, ErrDatabase, "", err, args...)
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

// Log the error using the default logger.
func (err *Error) Log() {
	logger.Error(err.Error())
}
