package jsonerror

import (
	"fmt"
	"reflect"
	"sync"
	"time"
)

// ErrorCollection can be configured to allow duplicates.
type DuplicatationOptions int

const (
	AllowDuplicates                 DuplicatationOptions = 0
	RejectDuplicatesIgnoreTimestamp DuplicatationOptions = 1 //Ignore timestamp information in JE struct
	RejectDuplicates                DuplicatationOptions = 2
)

// DefaultErrorFormatter represents the default formatter for displaying the collection
// of errors.
var DefaultErrorFormatter = func(i int, err error, str *string) {
	*str = fmt.Sprintf("%s\n%d:%s", *str, i, err.Error())
}

// ErrorCollection allows multiple errors to be accumulated and then returned as a single error.
// ErrorCollection can be safely used by concurrent goroutines.
type ErrorCollection struct {
	DuplicatationOptions DuplicatationOptions
	Errors               []error
	Formatter            func(i int, err error, str *string)
	lock                 sync.RWMutex
}

// Creates a new empty ErrorCollection.
// When `dup` is set, any duplicate error message is discarded
// and not appended to the collection
func NewErrorCollection(dup ...DuplicatationOptions) *ErrorCollection {
	ec := &ErrorCollection{}
	ec.Errors = []error{}
	ec.Formatter = DefaultErrorFormatter
	if len(dup) != 0 {
		ec.DuplicatationOptions = dup[0]
	}
	return ec
}

// Append an error to the error collection without locking
func (ec *ErrorCollection) addError(err error) {
	
	if err == nil {
		return
	}
	
	if ec.DuplicatationOptions != AllowDuplicates {
		//Don't append if err is a duplicate
		for i, containedErr := range ec.Errors {

			var je1 *JE
			var je2 *JE

			s, ok := err.(JE)
			if ok {
				je1 = &s
			} else {
				s, ok := err.(*JE)
				if ok {
					je1 = s
				}
			}

			_, ok = containedErr.(JE)
			if ok {
				t := (ec.Errors[i]).(JE)
				je2 = &t
			} else {
				_, ok := containedErr.(*JE)
				if ok {
					je2 = (ec.Errors[i]).(*JE)
				}
			}

			if je1 != nil && je2 != nil {
				//Don't use Reflection since both are JE structs
				if (*je1).Code == (*je2).Code && (*je1).Domain == (*je2).Domain && (*je1).error == (*je2).error && (*je1).message == (*je2).message {
					if ec.DuplicatationOptions == RejectDuplicates {
						if (*je1).time.Equal((*je2).time) {
							//Both JE structs are 100% identical including timestamp
							return
						}
					} else {
						//We don't care about timestamps
						return
					}
				}
			} else {
				//Use Reflection
				if reflect.DeepEqual(containedErr, err) {
					return
				}
			}
		}
	}
	ec.Errors = append(ec.Errors, err)
}

// AddError appends an error to the error collection.
// It is safe to use from multiple concurrent goroutines.
func (ec *ErrorCollection) AddError(err error) {
	ec.lock.Lock()
	defer ec.lock.Unlock()

	ec.addError(err)
}

// AddErrors appends multiple errors to the error collection.
// It is safe to use from multiple concurrent goroutines.
func (ec *ErrorCollection) AddErrors(errs ...error) {
	ec.lock.Lock()
	defer ec.lock.Unlock()

	for _, err := range errs {
		ec.addError(err)
	}
}

// AddErrorCollection appends an entire ErrorCollection to the receiver error collection.
// It is safe to use from multiple concurrent goroutines.
func (ec *ErrorCollection) AddErrorCollection(errs *ErrorCollection) {
	ec.lock.Lock()
	defer ec.lock.Unlock()

	for _, err := range errs.Errors {
		ec.addError(err)
	}
}

// Error return a list of all contained errors.
// The output can be formatted by setting a custom Formatter.
func (ec *ErrorCollection) Error() string {
	if ec.Formatter == nil {
		return ""
	}

	ec.lock.RLock()
	defer ec.lock.RUnlock()
	str := ""
	for i, err := range ec.Errors {
		if ec.Formatter != nil {
			ec.Formatter(i, err, &str)
		}
	}
	return str
}

// IsNil returns whether an error is nil or not.
// It can be used with ErrorCollection or generic errors
func IsNil(err error) bool {
	switch v := err.(type) {
	case *ErrorCollection:
		if len(v.Errors) == 0 {
			return true
		} else {
			return false
		}
	default:
		if err == nil {
			return true
		} else {
			return false
		}
	}
}

// JE allows errors to contain Code, Domain, Error and Message information.
// Only Code and Domain are exported so that once a JE struct is created, the key elements are static.
type JE struct {
	Code        int
	Domain      string
	error       string
	message     string
	time        time.Time //Displayed as Unix timestamp (number of nanoseconds elapsed since January 1, 1970 UTC)
	DisplayTime bool
}

// New creates a new JE struct.
// Domain is optional but can be at most 1 string.
func New(code int, error string, message string, domain ...string) JE {
	j := JE{Code: code, error: error, message: message, time: time.Now().UTC()}
	if len(domain) != 0 {
		j.Domain = domain[0]
	}
	return j
}

// NewAndDisplayTime creates a new JE struct and configures it to display the timestamp.
// Domain is optional but can be at most 1 string.
func NewAndDisplayTime(code int, error string, message string, domain ...string) JE {
	j := JE{Code: code, error: error, message: message, time: time.Now().UTC(), DisplayTime: true}
	if len(domain) != 0 {
		j.Domain = domain[0]
	}
	return j
}

// Error generates a string that neatly formats the contents of the JE struct.
// JSONError satisfies the error interface. Useful with panic().
func (j JE) Error() string {
	finalString := fmt.Sprintf("[code]: %d", j.Code)

	if j.error != "" {
		finalString = finalString + fmt.Sprintf(" [error]: %s", j.error)
	}

	if j.message != "" {
		finalString = finalString + fmt.Sprintf(" [message]: %s", j.message)
	}

	if j.Domain != "" {
		finalString = finalString + fmt.Sprintf(" [domain]: %s", j.Domain)
	}

	if j.DisplayTime {
		finalString = finalString + fmt.Sprintf(" [time]: %d", j.time.UnixNano())
	}

	return finalString
}

//Return the time the JE struct was created
func (j JE) Time() time.Time {
	return j.time
}

//For use with package: "gopkg.in/unrolled/render.v1".
//Can easily output properly formatted JSON error messages for REST API services.
func (j JE) Render() map[string]string {

	if j.error == "" {
		if j.message == "" {
			return map[string]string{"code": fmt.Sprintf("%d", j.Code)}
		} else {
			return map[string]string{"code": fmt.Sprintf("%d", j.Code), "message": j.message}
		}
	} else {
		if j.message == "" {
			return map[string]string{"code": fmt.Sprintf("%d", j.Code), "error": j.error}
		} else {
			return map[string]string{"code": fmt.Sprintf("%d", j.Code), "error": j.error, "message": j.message}
		}
	}
}
