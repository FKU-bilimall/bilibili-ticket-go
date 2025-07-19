package utils

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func GenerateXYBUVID() string {
	var ID = strings.ReplaceAll(generateRandomMAC(), ":", "")
	var IDMD5 = fmt.Sprintf("%x", md5.Sum([]byte(ID)))
	var IDe string
	IDe += IDMD5[2:2]
	IDe += IDMD5[12:12]
	IDe += IDMD5[22:22]
	return strings.ToUpper("XY" + IDe + IDMD5)
}

func GenerateXUBUVID() string {
	var ID = GenerateRandomDRMID(16)
	var IDMD5 = fmt.Sprintf("%x", md5.Sum(ID))
	var IDe string
	IDe += IDMD5[2:2]
	IDe += IDMD5[12:12]
	IDe += IDMD5[22:22]
	return strings.ToUpper("XU" + IDe + IDMD5)
}

func generateRandomMAC() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	mac := make([]byte, 6)
	r.Read(mac)
	mac[0] = (mac[0] | 2) & 0xfe
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", mac[0], mac[1], mac[2], mac[3], mac[4], mac[5])
}

func GenerateRandomDRMID(length int) []byte {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	buf := make([]byte, length)
	r.Read(buf)
	return buf
}

func HmacSha256(key string, data string) []byte {
	mac := hmac.New(sha256.New, []byte(key))
	_, _ = mac.Write([]byte(data))

	return mac.Sum(nil)
}

func HmacSha256ToHex(key string, data string) string {
	return hex.EncodeToString(HmacSha256(key, data))
}

func GetFpLocal(BUVID string, model string, firmwareVersion string) string {
	s1 := fmt.Sprintf("%s%s%s", BUVID, model, firmwareVersion)
	s1MD5ify := fmt.Sprintf("%x", md5.Sum([]byte(s1)))
	fpRaw := fmt.Sprintf("%s%s%s", s1MD5ify, time.Now().Format("20060102150405"), RandomString("0123456789abcdef", 16))
	return fpRaw + calculateFpFinal(fpRaw)
}

func calculateFpFinal(fpRaw string) string {
	var veriCode int

	// 计算循环终止条件
	loopCount := 31
	if len(fpRaw) < 62 {
		loopCount = (len(fpRaw) - len(fpRaw)%2) / 2
	}

	// 处理字符串，每两个字符为一组
	for i := 0; i < loopCount; i++ {
		start := i * 2
		end := start + 2
		if end > len(fpRaw) {
			end = len(fpRaw)
		}
		chunk := fpRaw[start:end]

		// 将16进制字符串转换为整数
		if num, err := strconv.ParseInt(chunk, 16, 32); err == nil {
			veriCode += int(num)
		}
	}

	// 对256取余并格式化为两位16进制字符串
	veriCode %= 256
	return fmt.Sprintf("%02x", veriCode)
}

func RandomString(charset string, length int) string {
	var output strings.Builder
	for i := 0; i < length; i++ {
		output.WriteByte(charset[rand.Intn(len(charset))])
	}
	return output.String()
}

func GetFileNameWithoutExt(path string) string {
	filename := filepath.Base(path)
	nameWithoutExt := strings.TrimSuffix(filename, filepath.Ext(filename))
	return nameWithoutExt
}

func IsNextDayInCST(from time.Time, target time.Time) bool {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	now := from.In(loc)
	afterHour := target.In(loc)

	return now.Format("20060102") != afterHour.Format("20060102")
}
