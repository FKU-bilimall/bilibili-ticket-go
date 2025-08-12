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

type BilibiliAPIVoucherError struct {
	Voucher string
}

func NewBilibiliAPIVoucherError(voucher string) *BilibiliAPIVoucherError {
	return &BilibiliAPIVoucherError{
		Voucher: voucher,
	}
}

func (bav *BilibiliAPIVoucherError) Error() string {
	return fmt.Sprintf("Need voucher: %s", bav.Voucher)
}
