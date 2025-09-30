package captcha

/*
#cgo linux,amd64   LDFLAGS: -L${SRCDIR}/biliTicker/target/release -L${SRCDIR}/biliTicker/target/x86_64-unknown-linux-gnu/release -lbili_ticket_gt -lm -ldl
#cgo linux,arm64   LDFLAGS: -L${SRCDIR}/biliTicker/target/release -L${SRCDIR}/biliTicker/target/aarch64-unknown-linux-gnu/release -lbili_ticket_gt -lm -ldl
#cgo darwin,amd64  LDFLAGS: -L${SRCDIR}/biliTicker/target/release -L${SRCDIR}/biliTicker/target/x86_64-apple-darwin/release -lbili_ticket_gt -lm -ldl
#cgo darwin,arm64  LDFLAGS: -L${SRCDIR}/biliTicker/target/release -L${SRCDIR}/biliTicker/target/aarch64-apple-darwin/release -lbili_ticket_gt -lm -ldl
#cgo windows,amd64 LDFLAGS: -L${SRCDIR}/biliTicker/target/release -L${SRCDIR}/biliTicker/target/x86_64-pc-windows-msvc/release -lbili_ticket_gt -lm
#cgo windows,arm64 LDFLAGS: -L${SRCDIR}/biliTicker/target/release -L${SRCDIR}/biliTicker/target/aarch64-pc-windows-msvc/release -lbili_ticket_gt -lm

#cgo CFLAGS: -I${SRCDIR}/biliTicker

#include "bindings.h"
*/
import "C"
import (
	"bilibili-ticket-go/models/enums"
	"bilibili-ticket-go/models/errors"
	"strings"
	"time"
	"unsafe"
)

// C Definitions
type cArgsBundle C.ArgsBundle
type cGeetestResult C.GeetestResult
type cGeetestCS C.GeetestCS

type GeetestCS struct {
	s string
	c []byte
}

type NewCSArgs struct {
	c  []byte
	s  string
	s1 string
	s2 string
	s3 string
	s4 string
}
type Click struct {
	isDestroyed bool
	inner       *C.ClickFFI
	gt          string
	challenge   string
}

func NewClick(gt, challenge string) *Click {
	return &Click{
		isDestroyed: false,
		inner:       C.new_click(),
		gt:          gt,
		challenge:   challenge,
	}
}

func (c *Click) GetCS(w string) (error, *GeetestCS) {
	if c.isDestroyed {
		return errors.NewCaptchaInstanceDestroyedError(), nil
	}
	cgtPtr := C.CString(c.gt)
	cchPtr := C.CString(c.challenge)
	defer C.free(unsafe.Pointer(cgtPtr))
	defer C.free(unsafe.Pointer(cchPtr))
	wPtr := C.CString(w)
	if w == "" {
		C.free(unsafe.Pointer(wPtr))
		wPtr = nil
	} else {
		defer C.free(unsafe.Pointer(wPtr))
	}
	geetestCS := C.click_get_c_s(c.inner, cgtPtr, cchPtr, wPtr)
	defer C.free_return_value(geetestCS, 3)
	if geetestCS.code == 1 {
		msg := C.GoString(geetestCS.message)
		return errors.NewCaptchaValidationError(msg), nil
	}
	cs := (*cGeetestCS)(geetestCS.data)
	bytes := C.GoBytes(unsafe.Pointer(cs.c_ptr), C.int(cs.c_len))
	s := C.GoString(cs.s)
	return nil, &GeetestCS{
		s: s,
		c: bytes,
	}
}

func (c *Click) GetType(w string) (error, enums.CaptchaType) {
	if c.isDestroyed {
		return errors.NewCaptchaInstanceDestroyedError(), enums.Unknown
	}
	cgtPtr := C.CString(c.gt)
	cchPtr := C.CString(c.challenge)
	defer C.free(unsafe.Pointer(cchPtr))
	defer C.free(unsafe.Pointer(cgtPtr))
	wPtr := C.CString(w)
	if w == "" {
		C.free(unsafe.Pointer(wPtr))
		wPtr = nil
	} else {
		defer C.free(unsafe.Pointer(wPtr))
	}
	typeRV := C.click_get_type(c.inner, cgtPtr, cchPtr, nil)
	defer C.free_return_value(typeRV, 0)
	if typeRV.code == 1 {
		msg := C.GoString(typeRV.message)
		return errors.NewCaptchaValidationError(msg), enums.Unknown
	}
	typeStr := strings.ToLower(C.GoString((*C.char)(typeRV.data)))
	var typeInt enums.CaptchaType
	switch typeStr {
	case "click":
		typeInt = enums.Click
	case "slide":
		typeInt = enums.Slide
	default:
		typeInt = enums.Unknown
	}
	return nil, typeInt
}

func (c *Click) FreeClick() {
	c.isDestroyed = true
	C.free_click(c.inner)
}

func (c *Click) GetNewCSArgs() (error, *NewCSArgs) {
	if c.isDestroyed {
		return errors.NewCaptchaInstanceDestroyedError(), nil
	}
	cgtPtr := C.CString(c.gt)
	cchPtr := C.CString(c.challenge)
	defer C.free(unsafe.Pointer(cgtPtr))
	defer C.free(unsafe.Pointer(cchPtr))
	newBundledArgs := C.click_get_new_c_s_args(c.inner, cgtPtr, cchPtr)
	defer C.free_return_value(newBundledArgs, 4)
	if newBundledArgs.code == 1 {
		msg := C.GoString(newBundledArgs.message)
		return errors.NewCaptchaValidationError(msg), nil
	}
	deboxedArgs := (*cArgsBundle)(newBundledArgs.data)
	bytes := C.GoBytes(unsafe.Pointer(deboxedArgs.c_ptr), C.int(deboxedArgs.c_len))
	s := C.GoString(deboxedArgs.s)
	s2 := C.GoString(deboxedArgs.full_bg_url)
	s3 := C.GoString(deboxedArgs.miss_bg_url)
	s4 := C.GoString(deboxedArgs.slider_url)
	s1 := C.GoString(deboxedArgs.new_challenge)
	return nil, &NewCSArgs{
		c:  bytes,
		s:  s,
		s1: s1,
		s2: s2,
		s3: s3,
		s4: s4,
	}
}

func (c *Click) CalculateKey(args *NewCSArgs) (error, string) {
	if c.isDestroyed {
		return errors.NewCaptchaInstanceDestroyedError(), ""
	}
	cgtPtr := C.CString(c.gt)
	cchPtr := C.CString(c.challenge)
	cncPtr := C.CString(args.s1)
	fbuPtr := C.CString(args.s2)
	mbuPtr := C.CString(args.s3)
	suPtr := C.CString(args.s4)
	defer C.free(unsafe.Pointer(cchPtr))
	defer C.free(unsafe.Pointer(cgtPtr))
	defer C.free(unsafe.Pointer(cncPtr))
	defer C.free(unsafe.Pointer(fbuPtr))
	defer C.free(unsafe.Pointer(mbuPtr))
	defer C.free(unsafe.Pointer(suPtr))
	keys := C.click_calculate_key(c.inner, cncPtr)
	defer C.free_return_value(keys, 0)
	if keys.code == 1 {
		msg := C.GoString(keys.message)
		return errors.NewCaptchaValidationError(msg), ""
	}
	return nil, C.GoString((*C.char)(keys.data))
}

func (c *Click) GenerateW(key string, args *NewCSArgs) (error, string) {
	if c.isDestroyed {
		return errors.NewCaptchaInstanceDestroyedError(), ""
	}
	cgtPtr := C.CString(c.gt)
	cchPtr := C.CString(c.challenge)
	defer C.free(unsafe.Pointer(cgtPtr))
	defer C.free(unsafe.Pointer(cchPtr))
	cArr := C.CBytes(args.c)
	cLen := C.int(len(args.c))
	defer C.free(cArr)
	sPtr := C.CString(args.s)
	defer C.free(unsafe.Pointer(sPtr))
	generate := C.click_generate_w(c.inner, C.CString(key), cgtPtr, cchPtr, (*C.uint8_t)(cArr), C.uintptr_t(cLen), sPtr)
	defer C.free_return_value(generate, 0)
	if generate.code == 1 {
		msg := C.GoString(generate.message)
		return errors.NewCaptchaValidationError(msg), ""
	}
	return nil, C.GoString((*C.char)(generate.data))
}

func (c *Click) Verify(w string) (error, string) {
	if c.isDestroyed {
		return errors.NewCaptchaInstanceDestroyedError(), ""
	}
	cgtPtr := C.CString(c.gt)
	cchPtr := C.CString(c.challenge)
	wPtr := C.CString(w)
	defer C.free(unsafe.Pointer(wPtr))
	defer C.free(unsafe.Pointer(cgtPtr))
	defer C.free(unsafe.Pointer(cchPtr))
	last := C.click_verify(c.inner, cgtPtr, cchPtr, wPtr)
	defer C.free_return_value(last, 2)
	if last.code == 1 {
		msg := C.GoString(last.message)
		return errors.NewCaptchaValidationError(msg), ""
	}
	res := (*cGeetestResult)(last.data)
	resMsg := C.GoString(res.message)
	if resMsg == "" {
		return errors.NewCaptchaValidationError(resMsg), ""
	}
	validate := C.GoString(res.validate)
	return nil, validate
}

func (c *Click) Solve() (error, string) {
	err, _ := c.GetCS("")
	if err != nil {
		return err, ""
	}
	err, tp := c.GetType("")
	if err != nil {
		return err, ""
	}
	if tp != enums.Click {
		return errors.NewCaptchaTypeMismatchError(enums.Click.String(), tp.String()), ""
	}
	err, args := c.GetNewCSArgs()
	if err != nil {
		return err, ""
	}
	beforeCalcTime := time.Now()
	err, key := c.CalculateKey(args)
	if err != nil {
		return err, ""
	}
	err, w := c.GenerateW(key, args)
	if err != nil {
		return err, ""
	}
	usedTime := time.Since(beforeCalcTime)
	if usedTime < 2*time.Second {
		time.Sleep(2*time.Second - usedTime)
	}
	err, s := c.Verify(w)
	if err != nil {
		return err, ""
	}
	return nil, s
}

type Slide struct {
	isDestroyed bool
	inner       *C.SlideFFI
	gt          string
	challenge   string
}

func NewSlide(gt, challenge string) *Slide {
	return &Slide{
		isDestroyed: false,
		inner:       C.new_slide(),
		gt:          gt,
		challenge:   challenge,
	}
}

func (s *Slide) GetCS(w string) (error, *GeetestCS) {
	if s.isDestroyed {
		return errors.NewCaptchaInstanceDestroyedError(), nil
	}
	cgtPtr := C.CString(s.gt)
	cchPtr := C.CString(s.challenge)
	defer C.free(unsafe.Pointer(cgtPtr))
	defer C.free(unsafe.Pointer(cchPtr))
	wPtr := C.CString(w)
	if w == "" {
		C.free(unsafe.Pointer(wPtr))
		wPtr = nil
	} else {
		defer C.free(unsafe.Pointer(wPtr))
	}
	geetestCS := C.slide_get_c_s(s.inner, cgtPtr, cchPtr, wPtr)
	defer C.free_return_value(geetestCS, 3)
	if geetestCS.code == 1 {
		msg := C.GoString(geetestCS.message)
		return errors.NewCaptchaValidationError(msg), nil
	}
	cs := (*cGeetestCS)(geetestCS.data)
	bytes := C.GoBytes(unsafe.Pointer(cs.c_ptr), C.int(cs.c_len))
	sval := C.GoString(cs.s)
	return nil, &GeetestCS{
		s: sval,
		c: bytes,
	}
}

func (s *Slide) GetType(w string) (error, enums.CaptchaType) {
	if s.isDestroyed {
		return errors.NewCaptchaInstanceDestroyedError(), enums.Unknown
	}
	cgtPtr := C.CString(s.gt)
	cchPtr := C.CString(s.challenge)
	defer C.free(unsafe.Pointer(cgtPtr))
	defer C.free(unsafe.Pointer(cchPtr))
	wPtr := C.CString(w)
	if w == "" {
		C.free(unsafe.Pointer(wPtr))
		wPtr = nil
	} else {
		defer C.free(unsafe.Pointer(wPtr))
	}
	typeRV := C.slide_get_type(s.inner, cgtPtr, cchPtr, nil)
	defer C.free_return_value(typeRV, 0)
	if typeRV.code == 1 {
		msg := C.GoString(typeRV.message)
		return errors.NewCaptchaValidationError(msg), enums.Unknown
	}
	typeStr := strings.ToLower(C.GoString((*C.char)(typeRV.data)))
	var typeInt enums.CaptchaType
	switch typeStr {
	case "slide":
		typeInt = enums.Slide
	case "click":
		typeInt = enums.Click
	default:
		typeInt = enums.Unknown
	}
	return nil, typeInt
}

func (s *Slide) FreeSlide() {
	s.isDestroyed = true
	C.free_slide(s.inner)
}

func (s *Slide) GetNewCSArgs() (error, *NewCSArgs) {
	if s.isDestroyed {
		return errors.NewCaptchaInstanceDestroyedError(), nil
	}
	cgtPtr := C.CString(s.gt)
	cchPtr := C.CString(s.challenge)
	defer C.free(unsafe.Pointer(cgtPtr))
	defer C.free(unsafe.Pointer(cchPtr))
	newBundledArgs := C.slide_get_new_c_s_args(s.inner, cgtPtr, cchPtr)
	defer C.free_return_value(newBundledArgs, 4)
	if newBundledArgs.code == 1 {
		msg := C.GoString(newBundledArgs.message)
		return errors.NewCaptchaValidationError(msg), nil
	}
	deboxedArgs := (*cArgsBundle)(newBundledArgs.data)
	bytes := C.GoBytes(unsafe.Pointer(deboxedArgs.c_ptr), C.int(deboxedArgs.c_len))
	sval := C.GoString(deboxedArgs.s)
	s2 := C.GoString(deboxedArgs.full_bg_url)
	s3 := C.GoString(deboxedArgs.miss_bg_url)
	s4 := C.GoString(deboxedArgs.slider_url)
	s1 := C.GoString(deboxedArgs.new_challenge)
	return nil, &NewCSArgs{
		c:  bytes,
		s:  sval,
		s1: s1,
		s2: s2,
		s3: s3,
		s4: s4,
	}
}

func (s *Slide) CalculateKey(args *NewCSArgs) (error, string) {
	if s.isDestroyed {
		return errors.NewCaptchaInstanceDestroyedError(), ""
	}
	cgtPtr := C.CString(s.gt)
	cchPtr := C.CString(s.challenge)
	cncPtr := C.CString(args.s1)
	fbuPtr := C.CString(args.s2)
	mbuPtr := C.CString(args.s3)
	suPtr := C.CString(args.s4)
	defer C.free(unsafe.Pointer(cgtPtr))
	defer C.free(unsafe.Pointer(cchPtr))
	defer C.free(unsafe.Pointer(cncPtr))
	defer C.free(unsafe.Pointer(fbuPtr))
	defer C.free(unsafe.Pointer(mbuPtr))
	defer C.free(unsafe.Pointer(suPtr))
	keys := C.slide_calculate_key(s.inner, cncPtr, fbuPtr, mbuPtr, suPtr)
	defer C.free_return_value(keys, 0)
	if keys.code == 1 {
		msg := C.GoString(keys.message)
		return errors.NewCaptchaValidationError(msg), ""
	}
	return nil, C.GoString((*C.char)(keys.data))
}

func (s *Slide) GenerateW(key string, args *NewCSArgs) (error, string) {
	if s.isDestroyed {
		return errors.NewCaptchaInstanceDestroyedError(), ""
	}
	cgtPtr := C.CString(s.gt)
	cchPtr := C.CString(s.challenge)
	defer C.free(unsafe.Pointer(cgtPtr))
	defer C.free(unsafe.Pointer(cchPtr))
	cArr := C.CBytes(args.c)
	cLen := C.int(len(args.c))
	defer C.free(cArr)
	sPtr := C.CString(args.s)
	defer C.free(unsafe.Pointer(sPtr))
	cncPtr := C.CString(args.s1)
	defer C.free(unsafe.Pointer(cncPtr))
	generate := C.slide_generate_w(s.inner, C.CString(key), cgtPtr, cncPtr, (*C.uint8_t)(cArr), C.uintptr_t(cLen), sPtr)
	defer C.free_return_value(generate, 0)
	if generate.code == 1 {
		msg := C.GoString(generate.message)
		return errors.NewCaptchaValidationError(msg), ""
	}
	return nil, C.GoString((*C.char)(generate.data))
}

func (s *Slide) Verify(w string) (error, string) {
	if s.isDestroyed {
		return errors.NewCaptchaInstanceDestroyedError(), ""
	}
	cgtPtr := C.CString(s.gt)
	cchPtr := C.CString(s.challenge)
	wPtr := C.CString(w)
	defer C.free(unsafe.Pointer(wPtr))
	defer C.free(unsafe.Pointer(cgtPtr))
	defer C.free(unsafe.Pointer(cchPtr))
	last := C.slide_verify(s.inner, cgtPtr, cchPtr, wPtr)
	defer C.free_return_value(last, 2)
	if last.code == 1 {
		msg := C.GoString(last.message)
		return errors.NewCaptchaValidationError(msg), ""
	}
	res := (*cGeetestResult)(last.data)
	resMsg := C.GoString(res.message)
	if resMsg == "" {
		return errors.NewCaptchaValidationError(resMsg), ""
	}
	validate := C.GoString(res.validate)
	return nil, validate
}

func (s *Slide) Solve() (error, string) {
	err, _ := s.GetCS("")
	if err != nil {
		return err, ""
	}
	err, tp := s.GetType("")
	if err != nil {
		return err, ""
	}
	if tp != enums.Click {
		return errors.NewCaptchaTypeMismatchError(enums.Click.String(), tp.String()), ""
	}
	err, args := s.GetNewCSArgs()
	if err != nil {
		return err, ""
	}
	err, key := s.CalculateKey(args)
	if err != nil {
		return err, ""
	}
	err, w := s.GenerateW(key, args)
	if err != nil {
		return err, ""
	}
	err, v := s.Verify(w)
	if err != nil {
		return err, ""
	}
	return nil, v
}
