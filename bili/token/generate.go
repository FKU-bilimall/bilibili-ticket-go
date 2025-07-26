package token

import (
	"encoding/base64"
	"math/rand/v2"
	"time"
)

type CTokenGenerator struct {
	begin          time.Time
	generateCounts int
}

func NewCTokenGenerator() *CTokenGenerator {
	return &CTokenGenerator{
		begin: time.Now(),
	}
}

type windowStats struct {
	TouchCount   uint16 // f
	VisibleCount uint16 // d
	UnloadCount  uint16 // p
	StaySeconds  uint16 // h
	SinceInitSec uint16 // v
	ScrollX      uint16 // m
	ScrollY      uint16 // y
	InnerWidth   uint16 // g
	InnerHeight  uint16 // b
	OuterWidth   uint16 // _
	OuterHeight  uint16 // w
	ScreenX      uint16 // A
	ScreenY      uint16 // x
	ScreenWidth  uint16 // C
	ScreenHeight uint16 // k
	AvailWidth   uint16 // E
}

func makeToken(w *windowStats) string {
	data := make([]byte, 16)
	// Begin mapping just like JS:
	// 0: TouchCount, 1: ScrollX, 2: VisibleCount, 3: ScrollY,
	// 4: InnerWidth, 5: UnloadCount, 6: InnerHeight, 7: OuterWidth,
	// 8: StaySeconds (2 bytes), 10: SinceInitSec (2 bytes),
	// 12: OuterHeight, 13: ScreenX, 14: ScreenY, 15: ScreenWidth
	mapping := [16]struct {
		Index  int
		Value  uint16
		Length int
	}{
		{0, w.TouchCount, 1},
		{1, w.ScrollX, 1},
		{2, w.VisibleCount, 1},
		{3, w.ScrollY, 1},
		{4, w.InnerWidth, 1},
		{5, w.UnloadCount, 1},
		{6, w.InnerHeight, 1},
		{7, w.OuterWidth, 1},
		{8, w.StaySeconds, 2},
		{10, w.SinceInitSec, 2},
		{12, w.OuterHeight, 1},
		{13, w.ScreenX, 1},
		{14, w.ScreenY, 1},
		{15, w.ScreenWidth, 1},
	}

	filled := make(map[int]bool)
	for _, field := range mapping {
		filled[field.Index] = true
		if field.Length == 1 {
			if field.Value > 255 {
				data[field.Index] = 255
			} else {
				data[field.Index] = uint8(field.Value)
			}
		} else if field.Length == 2 {
			if field.Value > 65535 {
				field.Value = 65535
			}
			data[field.Index] = uint8(field.Value >> 8)
			data[field.Index+1] = uint8(field.Value)
		}
	}
	// Handle 9, 11 (填充位)
	// JS: n.setUint8(o, 4 & k ? y : E)
	for _, i := range []int{9, 11} {
		if !filled[i] {
			if w.ScreenHeight&4 != 0 {
				if w.ScrollY > 255 {
					data[i] = 255
				} else {
					data[i] = uint8(w.ScrollY)
				}
			} else {
				if w.AvailWidth > 255 {
					data[i] = 255
				} else {
					data[i] = uint8(w.AvailWidth)
				}
			}
		}
	}
	// 此时data为16字节
	// toBinary类似：每char取u16，再转u8字流，然后btoa
	// 这里简化为直接base64编码即可（JS的转法是上一个细节，实际就是生成的二进制流base64出字符串）
	encoded := base64.StdEncoding.EncodeToString(data)
	return encoded
}

func (g *CTokenGenerator) GenerateCTokenPrepareStage() string {
	t := makeToken(&windowStats{
		TouchCount:   uint16(rand.IntN(7) + 3),
		VisibleCount: uint16(rand.IntN(13) + 3),
		UnloadCount:  uint16(g.generateCounts),
		StaySeconds:  uint16(time.Now().Sub(g.begin).Seconds()),
		SinceInitSec: 0,
		ScrollX:      0,
		ScrollY:      0,
		InnerWidth:   1578,
		InnerHeight:  690,
		OuterWidth:   1578,
		OuterHeight:  690,
		ScreenX:      1699,
		ScreenY:      834,
		ScreenWidth:  1699,
		ScreenHeight: 834,
		AvailWidth:   1578,
	})
	g.generateCounts++
	return t
}

func (g *CTokenGenerator) GenerateCTokenCreateStage(whenGenPToken time.Time) string {
	t := makeToken(&windowStats{
		TouchCount:   uint16(rand.IntN(7) + 3),
		VisibleCount: uint16(rand.IntN(13) + 3),
		UnloadCount:  uint16(g.generateCounts),
		StaySeconds:  uint16(time.Now().Sub(g.begin).Seconds()),
		SinceInitSec: uint16(time.Now().Sub(whenGenPToken).Seconds()),
		ScrollX:      0,
		ScrollY:      0,
		InnerWidth:   1578,
		InnerHeight:  690,
		OuterWidth:   1578,
		OuterHeight:  690,
		ScreenX:      1699,
		ScreenY:      834,
		ScreenWidth:  1699,
		ScreenHeight: 834,
		AvailWidth:   1578,
	})
	g.generateCounts++
	return t
}
