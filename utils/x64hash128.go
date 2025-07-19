package utils

import (
	"encoding/binary"
)

// murmurhash3X64128 实现了MurmurHash3的128位版本
type murmurhash3X64128 struct{}

// 64位加法操作
// JavaScript中用两个32位数组表示64位数，Go中直接使用uint64
func (m *murmurhash3X64128) x64Add(a, b uint64) uint64 {
	return a + b
}

// 64位乘法操作
func (m *murmurhash3X64128) x64Multiply(a, b uint64) uint64 {
	return a * b
}

// 64位循环左移
// 将位向左移动n位，溢出的位补到右边
func (m *murmurhash3X64128) x64Rotl(x uint64, n uint8) uint64 {
	n %= 64
	if n == 0 {
		return x
	}
	return (x << n) | (x >> (64 - n))
}

// 64位左移
// 普通左移，溢出位丢弃
func (m *murmurhash3X64128) x64LeftShift(x uint64, n uint8) uint64 {
	n %= 64
	if n == 0 {
		return x
	}
	return x << n
}

// 64位异或操作
func (m *murmurhash3X64128) x64Xor(a, b uint64) uint64 {
	return a ^ b
}

// 最终混合函数
// 用于确保哈希值的雪崩效应（输入的微小变化导致输出的巨大变化）
func (m *murmurhash3X64128) x64Fmix(k uint64) uint64 {
	k ^= k >> 33
	k *= 0xff51afd7ed558ccd
	k ^= k >> 33
	k *= 0xc4ceb9fe1a85ec53
	k ^= k >> 33
	return k
}

// MurmurX64Hash128 计算给定字符串的128位MurmurHash3值
// seed是种子值，用于生成不同的哈希序列
func MurmurX64Hash128(key string, seed uint32) (uint64, uint64) {
	data := []byte(key)
	nblocks := len(data) / 16

	h1 := uint64(seed)
	h2 := uint64(seed)

	// 常量，是算法的魔数
	c1 := uint64(0x87c37b91114253d5)
	c2 := uint64(0x4cf5ad432745937f)

	m := &murmurhash3X64128{}
	// 主循环：每次处理16字节（128位）
	for i := 0; i < nblocks; i++ {
		// 读取两个64位块
		k1 := binary.LittleEndian.Uint64(data[i*16:])
		k2 := binary.LittleEndian.Uint64(data[i*16+8:])

		// 第一个块的处理
		k1 *= c1
		k1 = m.x64Rotl(k1, 31)
		k1 *= c2
		h1 ^= k1

		h1 = m.x64Rotl(h1, 27)
		h1 += h2
		h1 = h1*5 + 0x52dce729

		// 第二个块的处理
		k2 *= c2
		k2 = m.x64Rotl(k2, 33)
		k2 *= c1
		h2 ^= k2

		h2 = m.x64Rotl(h2, 31)
		h2 += h1
		h2 = h2*5 + 0x38495ab5
	}

	// 处理剩余字节（尾部处理）
	tail := data[nblocks*16:]
	k1 := uint64(0)
	k2 := uint64(0)

	// 使用switch fall-through处理剩余字节
	switch len(tail) {
	case 15:
		k2 ^= uint64(tail[14]) << 48
		fallthrough
	case 14:
		k2 ^= uint64(tail[13]) << 40
		fallthrough
	case 13:
		k2 ^= uint64(tail[12]) << 32
		fallthrough
	case 12:
		k2 ^= uint64(tail[11]) << 24
		fallthrough
	case 11:
		k2 ^= uint64(tail[10]) << 16
		fallthrough
	case 10:
		k2 ^= uint64(tail[9]) << 8
		fallthrough
	case 9:
		k2 ^= uint64(tail[8])
		k2 *= c2
		k2 = m.x64Rotl(k2, 33)
		k2 *= c1
		h2 ^= k2
		fallthrough
	case 8:
		k1 ^= uint64(tail[7]) << 56
		fallthrough
	case 7:
		k1 ^= uint64(tail[6]) << 48
		fallthrough
	case 6:
		k1 ^= uint64(tail[5]) << 40
		fallthrough
	case 5:
		k1 ^= uint64(tail[4]) << 32
		fallthrough
	case 4:
		k1 ^= uint64(tail[3]) << 24
		fallthrough
	case 3:
		k1 ^= uint64(tail[2]) << 16
		fallthrough
	case 2:
		k1 ^= uint64(tail[1]) << 8
		fallthrough
	case 1:
		k1 ^= uint64(tail[0])
		k1 *= c1
		k1 = m.x64Rotl(k1, 31)
		k1 *= c2
		h1 ^= k1
	}

	// 最终化
	h1 ^= uint64(len(data))
	h2 ^= uint64(len(data))

	h1 += h2
	h2 += h1

	h1 = m.x64Fmix(h1)
	h2 = m.x64Fmix(h2)

	h1 += h2
	h2 += h1

	return h1, h2
}
