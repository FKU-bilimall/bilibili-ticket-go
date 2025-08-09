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

func makeToken(stats *windowStats) string {
	// 创建16字节缓冲区
	buf := make([]byte, 16)

	// 映射配置 - 使用结构体字段
	fields := map[int]struct {
		data   uint16
		length int
	}{
		0:  {stats.TouchCount, 1},
		1:  {stats.ScrollX, 1},
		2:  {stats.VisibleCount, 1},
		3:  {stats.ScrollY, 1},
		4:  {stats.InnerWidth, 1},
		5:  {stats.UnloadCount, 1},
		6:  {stats.InnerHeight, 1},
		7:  {stats.OuterWidth, 1},
		8:  {stats.StaySeconds, 2},
		10: {stats.SinceInitSec, 2},
		12: {stats.OuterHeight, 1},
		13: {stats.ScreenX, 1},
		14: {stats.ScreenY, 1},
		15: {stats.ScreenWidth, 1},
	}

	// 填充字节（对应JS的DataView操作）
	for o := 0; o < 16; o++ {
		if f, ok := fields[o]; ok {
			if f.length == 1 {
				if f.data > 255 {
					buf[o] = 255
				} else {
					buf[o] = byte(f.data)
				}
			} else {
				val := f.data
				if val > 65535 {
					val = 65535
				}
				// 注意：JS中setUint16默认是使用平台字节序（通常是小端序）
				buf[o] = byte(val & 0xFF)          // 低字节
				buf[o+1] = byte((val >> 8) & 0xFF) // 高字节
				o++
			}
		} else {
			// 与屏幕高度和可用宽度有关的逻辑
			if stats.ScreenHeight&4 != 0 {
				buf[o] = byte(stats.ScrollY)
			} else {
				buf[o] = byte(stats.AvailWidth)
			}
		}
	}

	// 正确模拟JS的toBinary：每个字节转为Uint16后平坦化为字节数组
	result := make([]byte, 32)
	for i := 0; i < 16; i++ {
		// 模拟 JS 中 Uint16Array (小端序) 的内存布局
		result[i*2] = buf[i]
		result[i*2+1] = 0x00
	}

	return base64.StdEncoding.EncodeToString(result)
}

func (g *CTokenGenerator) GenerateTokenPrepareStage() string {
	t := makeToken(&windowStats{
		TouchCount:   uint16(rand.IntN(7) + 3),
		VisibleCount: uint16(rand.IntN(2) + 3),
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

func (g *CTokenGenerator) GenerateTokenCreateStage(whenGenPToken time.Time) string {
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

func (g *CTokenGenerator) IsHotProject() bool {
	return true
}
