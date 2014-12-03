package jsonerror

import (
	"fmt"
)

type JSONError struct {
	code    int
	error   string
	message string
}

func New(code int, error string, message string) *JSONError {
	return &JSONError{code: code, error: error, message: message}
}

func (self JSONError) Error() string {
	finalString := fmt.Sprintf("code: %d", self.code)

	if self.error != "" {
		finalString = finalString + fmt.Sprintf("error: %s", self.error)
	}

	if self.message != "" {
		finalString = finalString + fmt.Sprintf("message: %s", self.message)
	}

	return finalString
}

func (self JSONError) Render() map[string]string {

	if self.error == "" {
		if self.message == "" {
			return map[string]string{"code": fmt.Sprintf("%d", self.code)}
		} else {
			return map[string]string{"code": fmt.Sprintf("%d", self.code), "message": self.message}
		}
	} else {
		if self.message == "" {
			return map[string]string{"code": fmt.Sprintf("%d", self.code), "error": self.error}
		} else {
			return map[string]string{"code": fmt.Sprintf("%d", self.code), "error": self.error, "message": self.message}
		}
	}
}
