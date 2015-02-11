package jsonerror

import (
	"fmt"
)

type JSONError struct {
	Code    int
	error   string
	message string
	Domain  string
}

func New(code int, error string, message string, domain ...string) *JSONError {
	if len(domain) == 0 {
		return &JSONError{Code: code, error: error, message: message}
	} else {
		return &JSONError{Code: code, error: error, message: message, Domain: domain[0]}
	}
}

func (self JSONError) Error() string {
	finalString := fmt.Sprintf("code: %d", self.Code)

	if self.error != "" {
		finalString = finalString + fmt.Sprintf(" error: %s", self.error)
	}

	if self.message != "" {
		finalString = finalString + fmt.Sprintf(" message: %s", self.message)
	}

	if self.Domain != "" {
		finalString = finalString + fmt.Sprintf(" domain: %s", self.Domain)
	}

	return finalString
}

//For use with package: gopkg.in/unrolled/render.v1
//Can easily output properly formatted JSON error messages for REST API.
func (self JSONError) Render() map[string]string {

	if self.error == "" {
		if self.message == "" {
			return map[string]string{"code": fmt.Sprintf("%d", self.Code)}
		} else {
			return map[string]string{"code": fmt.Sprintf("%d", self.Code), "message": self.message}
		}
	} else {
		if self.message == "" {
			return map[string]string{"code": fmt.Sprintf("%d", self.Code), "error": self.error}
		} else {
			return map[string]string{"code": fmt.Sprintf("%d", self.Code), "error": self.error, "message": self.message}
		}
	}
}
