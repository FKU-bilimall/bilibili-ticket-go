package utils

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
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

func HmacSha256(key string, data string) []byte {
	mac := hmac.New(sha256.New, []byte(key))
	_, _ = mac.Write([]byte(data))

	return mac.Sum(nil)
}

func HmacSha256ToHex(key string, data string) string {
	return hex.EncodeToString(HmacSha256(key, data))
}
