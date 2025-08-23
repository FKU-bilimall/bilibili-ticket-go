package utils

import (
	"crypto/md5"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Deprecated: Use GenerateXUBUVID instead.
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

var InfocDigitMap = []string{
	"1", "2", "3", "4", "5", "6", "7", "8", "9", "A", "B", "C", "D", "E", "F", "10",
}

func GenerateUUIDInfoc() string {
	t := time.Now().UnixMilli() % 100000
	return strings.Join([]string{
		randomChoice([]int{8, 4, 4, 4, 12}, "-", InfocDigitMap),
		fmt.Sprintf("%05d", t),
		"infoc",
	}, "")
}

func IsTicketOnSale(flag int) bool {
	switch flag {
	case 2:
		// 预售中
		return true
	case 3:
		// 已停售
		return false
	case 4:
		// 已售罄
		return true
	case 5:
		// 不可售
		return false
	default:
		return false
	}
}
