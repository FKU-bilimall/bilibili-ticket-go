package utils

import (
	"image"

	"github.com/skip2/go-qrcode"
)

func GetQRCode(content string, isFlat bool) ([]string, image.Image) {
	q, err := qrcode.New(content, qrcode.Low)
	if err != nil {
		panic(err)
	}
	q.DisableBorder = true
	bits := q.Bitmap()
	var ASCIIQR []string
	if isFlat {
		for j := 0; j < len(bits[0]); j++ {
			var line = ""
			for i := 0; i < len(bits); i++ {
				if bits[j][i] {
					line += "██"
				} else {
					line += "  "
				}
			}
			ASCIIQR = append(ASCIIQR, line)
		}
	} else {
		for j := 0; j < len(bits[0]); j += 2 {
			var line = ""
			for i := 0; i < len(bits); i += 2 {
				var lt = bits[j][i]
				var lb, rt, rb bool
				if i+2 >= len(bits[0]) {
					rt = false
				} else {
					rt = bits[j][i+1]
				}
				if j+2 >= len(bits) {
					lb = false
				} else {
					lb = bits[j+1][i]
				}
				if j+2 >= len(bits) || i+2 >= len(bits[0]) {
					rb = false
				} else {
					rb = bits[j+1][i+1]
				}
				if lt && lb && rt && rb {
					line += "██"
				} else if lt && lb && rt && !rb {
					line += "█▀"
				} else if lt && lb && !rt && rb {
					line += "█▄"
				} else if lt && !lb && rt && rb {
					line += "▀█"
				} else if !lt && lb && rt && rb {
					line += "▄█"
				} else if lt && lb && !rt && !rb {
					line += "█ "
				} else if lt && !lb && rt && !rb {
					line += "▀▀"
				} else if lt && !lb && !rt && rb {
					line += "▀▄"
				} else if !lt && lb && rt && !rb {
					line += "▄▀"
				} else if !lt && lb && !rt && rb {
					line += "▄▄"
				} else if !lt && !lb && rt && rb {
					line += " █"
				} else if lt && !lb && !rt && !rb {
					line += "▀ "
				} else if !lt && lb && !rt && !rb {
					line += "▄ "
				} else if !lt && !lb && rt && !rb {
					line += " ▀"
				} else if !lt && !lb && !rt && rb {
					line += " ▄"
				} else {
					line += "  "
				}
			}
			ASCIIQR = append(ASCIIQR, line)
		}
	}
	q.DisableBorder = false
	return ASCIIQR, q.Image(256)
}
