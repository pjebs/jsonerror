package jsonerror

import (
	"fmt"
	"reflect"
	"sync"
	"time"
)

// ErrorCollection can be configured to allow duplicates.
type DuplicatesOptions int

const (
	AllowDuplicates                 DuplicatesOptions = 0
	RejectDuplicatesIgnoreTimestamp DuplicatesOptions = 1 //Ignore timestamp information in JE struct
	RejectDuplicates                DuplicatesOptions = 2
)

// DefaultErrorFormatter represents the default formatter for displaying the collection
// of errors.
var DefaultErrorFormatter = func(i int, err error, str *string) {
	*str = fmt.Sprintf("%s\n%d:%s", *str, i, err.Error())
}

// ErrorCollection allows multiple errors to be accumulated and then returned as a single error.
// ErrorCollection can be safely used by concurrent go-routines.
type ErrorCollection struct {
	RemoveDuplicates DuplicatesOptions
	Errors           []error
	Formatter        func(i int, err error, str *string)
	lock             sync.RWMutex
}

// Creates a new empty ErrorCollection.
// When removeDuplicates is set, any duplicate error messages are discarded
// and not appended to the collection
func NewErrorCollection(removeDuplicates ...DuplicatesOptions) *ErrorCollection {
	ec := &ErrorCollection{}
	ec.Errors = []error{}
	ec.Formatter = DefaultErrorFormatter
	if len(removeDuplicates) != 0 {
		ec.RemoveDuplicates = removeDuplicates[0]
	}
	return ec
}

// Append an error to the error collection without locking
func (ec *ErrorCollection) addError(err error) {
	if ec.RemoveDuplicates != AllowDuplicates {
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
					if ec.RemoveDuplicates == RejectDuplicates {
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

// Append an error to the error collection
func (ec *ErrorCollection) AddError(err error) {
	ec.lock.Lock()
	defer ec.lock.Unlock()

	ec.addError(err)
}

// Append multiple errors to the error collection
func (ec *ErrorCollection) AddErrors(errs ...error) {
	ec.lock.Lock()
	defer ec.lock.Unlock()

	for _, err := range errs {
		ec.addError(err)
	}
}

// Append an entire ErrorCollection to the receiver error collection
func (ec *ErrorCollection) AddErrorCollection(errs *ErrorCollection) {
	ec.lock.Lock()
	defer ec.lock.Unlock()

	for _, err := range errs.Errors {
		ec.addError(err)
	}
}

// Return a list of all contained errors
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

//Struct that contains the Code, Domain, Error and Message.
//Only Code and Domain are exported to encourage usage of the New(...) method to set the message and error.
//JE is shorthand for JSONError.
type JE struct {
	Code    int
	Domain  string
	error   string
	message string
	time    time.Time
}

//Creates a new JE struct.
//Domain is optional but can be at most 1 string.
func New(code int, error string, message string, domain ...string) JE {
	j := JE{Code: code, error: error, message: message, time: time.Now().UTC()}
	if len(domain) != 0 {
		j.Domain = domain[0]
	}
	return j
}

//Generates a string that neatly formats the contents of JE struct.
//Useful with panic() because JSONError satisfies error interface.
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
