package utils

import (
	"strconv"
	"strings"

	"github.com/fatih/color"
)

func hexToColor(hex string) (*color.Color, error) {
	// 去除可能的 # 前缀
	hex = strings.TrimPrefix(hex, "#")

	// 解析RGB值
	r, err := strconv.ParseUint(hex[0:2], 16, 8)
	if err != nil {
		return nil, err
	}
	g, err := strconv.ParseUint(hex[2:4], 16, 8)
	if err != nil {
		return nil, err
	}
	b, err := strconv.ParseUint(hex[4:6], 16, 8)
	if err != nil {
		return nil, err
	}

	return color.RGB(int(r), int(g), int(b)), nil
}
