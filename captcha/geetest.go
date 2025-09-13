package captcha

/*
#cgo linux,amd64   LDFLAGS: -L${SRCDIR}/biliTicker/target/release -L${SRCDIR}/biliTicker/target/x86_64-unknown-linux-gnu/release -lbili_ticket_gt -lm -ldl
#cgo linux,arm64   LDFLAGS: -L${SRCDIR}/biliTicker/target/release -L${SRCDIR}/biliTicker/target/aarch64-unknown-linux-gnu/release -lbili_ticket_gt -lm -ldl
#cgo darwin,amd64  LDFLAGS: -L${SRCDIR}/biliTicker/target/release -L${SRCDIR}/biliTicker/target/x86_64-apple-darwin/release -lbili_ticket_gt -lm -ldl
#cgo darwin,arm64  LDFLAGS: -L${SRCDIR}/biliTicker/target/release -L${SRCDIR}/biliTicker/target/aarch64-apple-darwin/release -lbili_ticket_gt -lm -ldl
#cgo windows,amd64 LDFLAGS: -L${SRCDIR}/biliTicker/target/release -L${SRCDIR}/biliTicker/target/aarch64-pc-windows-msvc/release -lbili_ticket_gt -lm
#cgo windows,arm64 LDFLAGS: -L${SRCDIR}/biliTicker/target/release -L${SRCDIR}/biliTicker/target/x86_64-pc-windows-msvc/release -lbili_ticket_gt -lm

#cgo CFLAGS: -I${SRCDIR}/biliTicker

#include "bindings.h"
*/
import "C"
import (
	"bilibili-ticket-go/models/errors"
	"time"
	"unsafe"
)

type ReturnValue C.ReturnValue

type ArgsBundle C.ArgsBundle

type GeetestResult C.GeetestResult

func SolveClick(challenge, gt string) (error, string) {
	clkPtr := C.new_click()
	defer C.free_click(clkPtr)
	cchPtr := C.CString(challenge)
	cgtPtr := C.CString(gt)
	defer C.free(unsafe.Pointer(cchPtr))
	defer C.free(unsafe.Pointer(cgtPtr))
	cs := C.click_get_c_s(clkPtr, cgtPtr, cchPtr, nil) //ignore
	C.free_return_value(cs, 3)
	rt := C.click_get_type(clkPtr, cgtPtr, cchPtr, nil)
	defer C.free_return_value(rt, 0)
	if rt.code == 1 {
		msg := C.GoString(rt.message)
		return errors.NewCaptchaValidationError(msg), ""
	}
	typeStr := C.GoString((*C.char)(rt.data))
	if typeStr != "click" {
		return errors.NewCaptchaTypeMismatchError("click", typeStr), ""
	}
	newBundledArgs := C.click_get_new_c_s_args(clkPtr, cgtPtr, cchPtr)
	defer C.free_return_value(newBundledArgs, 4)
	if newBundledArgs.code == 1 {
		msg := C.GoString(newBundledArgs.message)
		return errors.NewCaptchaValidationError(msg), ""
	}
	deboxedArgs := (*ArgsBundle)(newBundledArgs.data)
	beforeCalcTime := time.Now()
	keys := C.click_calculate_key(clkPtr, deboxedArgs.new_challenge)
	defer C.free_return_value(keys, 0)
	if keys.code == 1 {
		msg := C.GoString(keys.message)
		return errors.NewCaptchaValidationError(msg), ""
	}
	generate := C.click_generate_w(clkPtr, (*C.char)(keys.data), cgtPtr, cchPtr, deboxedArgs.c_ptr, deboxedArgs.c_len, deboxedArgs.s)
	defer C.free_return_value(generate, 0)
	if generate.code == 1 {
		msg := C.GoString(generate.message)
		return errors.NewCaptchaValidationError(msg), ""
	}
	usedTime := time.Since(beforeCalcTime)
	if usedTime < 2*time.Second {
		time.Sleep(2*time.Second - usedTime)
	}
	last := C.click_verify(clkPtr, cgtPtr, cchPtr, (*C.char)(generate.data))
	defer C.free_return_value(last, 2)
	if last.code == 1 {
		msg := C.GoString(last.message)
		return errors.NewCaptchaValidationError(msg), ""
	}
	res := (*GeetestResult)(last.data)
	resMsg := C.GoString(res.message)
	if resMsg == "" {
		return errors.NewCaptchaValidationError(resMsg), ""
	}
	validate := C.GoString(res.validate)
	return nil, validate
}

func SolveSlide(challenge, gt string) (error, string) {
	clkPtr := C.new_slide()
	defer C.free_slide(clkPtr)
	cchPtr := C.CString(challenge)
	cgtPtr := C.CString(gt)
	defer C.free(unsafe.Pointer(cchPtr))
	defer C.free(unsafe.Pointer(cgtPtr))
	cs := C.slide_get_c_s(clkPtr, cgtPtr, cchPtr, nil) //ignore
	C.free_return_value(cs, 3)
	rt := C.slide_get_type(clkPtr, cgtPtr, cchPtr, nil)
	defer C.free_return_value(rt, 0)
	if rt.code == 1 {
		msg := C.GoString(rt.message)
		return errors.NewCaptchaValidationError(msg), ""
	}
	typeStr := C.GoString((*C.char)(rt.data))
	if typeStr != "slide" {
		return errors.NewCaptchaTypeMismatchError("slide", typeStr), ""
	}
	newBundledArgs := C.slide_get_new_c_s_args(clkPtr, cgtPtr, cchPtr)
	defer C.free_return_value(newBundledArgs, 4)
	if newBundledArgs.code == 1 {
		msg := C.GoString(newBundledArgs.message)
		return errors.NewCaptchaValidationError(msg), ""
	}
	deboxedArgs := (*ArgsBundle)(newBundledArgs.data)
	keys := C.slide_calculate_key(clkPtr, deboxedArgs.new_challenge, deboxedArgs.full_bg_url, deboxedArgs.miss_bg_url, deboxedArgs.slider_url)
	defer C.free_return_value(keys, 0)
	if keys.code == 1 {
		msg := C.GoString(keys.message)
		return errors.NewCaptchaValidationError(msg), ""
	}
	generate := C.slide_generate_w(clkPtr, (*C.char)(keys.data), cgtPtr, deboxedArgs.new_challenge, deboxedArgs.c_ptr, deboxedArgs.c_len, deboxedArgs.s)
	defer C.free_return_value(generate, 0)
	if generate.code == 1 {
		msg := C.GoString(generate.message)
		return errors.NewCaptchaValidationError(msg), ""
	}
	last := C.slide_verify(clkPtr, cgtPtr, cchPtr, (*C.char)(generate.data))
	defer C.free_return_value(last, 2)
	if last.code == 1 {
		msg := C.GoString(last.message)
		return errors.NewCaptchaValidationError(msg), ""
	}
	res := (*GeetestResult)(last.data)
	resMsg := C.GoString(res.message)
	if resMsg == "" {
		return errors.NewCaptchaValidationError(resMsg), ""
	}
	validate := C.GoString(res.validate)
	return nil, validate
}
