package jsonerror

import (
	"fmt"
)

//Struct that contains the Code, Domain, Error and Message.
//Only Code and Domain are exported to encourage usage of the New(...) method to set the message and error.
//JE is shorthand for JSONError.
type JE struct {
	Code    int
	Domain  string
	error   string
	message string
}

//Creates a new JE struct.
//Domain is optional but can be at most 1 string.
func New(code int, error string, message string, domain ...string) JE {
	if len(domain) == 0 {
		return JE{Code: code, error: error, message: message}
	} else {
		return JE{Code: code, error: error, message: message, Domain: domain[0]}
	}
}

//Generates a string that neatly formats the contents of JE struct.
//Useful with panic() because JSONError satisfies error interface.
func (this JE) Error() string {
	finalString := fmt.Sprintf("[code]: %d", this.Code)

	if this.error != "" {
		finalString = finalString + fmt.Sprintf(" [error]: %s", this.error)
	}

	if this.message != "" {
		finalString = finalString + fmt.Sprintf(" [message]: %s", this.message)
	}

	if this.Domain != "" {
		finalString = finalString + fmt.Sprintf(" [domain]: %s", this.Domain)
	}

	return finalString
}

//For use with package: "gopkg.in/unrolled/render.v1".
//Can easily output properly formatted JSON error messages for REST API services.
func (this JE) Render() map[string]string {

	if this.error == "" {
		if this.message == "" {
			return map[string]string{"code": fmt.Sprintf("%d", this.Code)}
		} else {
			return map[string]string{"code": fmt.Sprintf("%d", this.Code), "message": this.message}
		}
	} else {
		if this.message == "" {
			return map[string]string{"code": fmt.Sprintf("%d", this.Code), "error": this.error}
		} else {
			return map[string]string{"code": fmt.Sprintf("%d", this.Code), "error": this.error, "message": this.message}
		}
	}
}
