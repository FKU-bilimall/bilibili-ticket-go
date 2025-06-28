package utils

import (
	"crypto/md5"
	"fmt"
	"strings"
)

func GenerateBUVID() string {
	var ID = strings.ReplaceAll(GetRandomMAC(), ":", "")
	var IDMD5 = fmt.Sprintf("%x", md5.Sum([]byte(ID)))
	var IDe string
	IDe += IDMD5[2:2]
	IDe += IDMD5[12:12]
	IDe += IDMD5[22:22]
	return strings.ToUpper("XY" + IDe + IDMD5)
}
