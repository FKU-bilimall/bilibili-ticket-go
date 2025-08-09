package errors

import "fmt"

type BilibiliAPIError struct {
	Code    int
	Message string
}

func NewBilibiliAPIError(code int, message string) *BilibiliAPIError {
	return &BilibiliAPIError{
		Code:    code,
		Message: message,
	}
}

func (bae *BilibiliAPIError) Error() string {
	return fmt.Sprintf("Response code is not 0, got: %d, message: %s", bae.Code, bae.Message)
}
