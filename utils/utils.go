package utils

import (
	"image/color"
	"math/rand"
)

func SortAsc(s []uint32, i, j int) bool {
	return s[i] < s[j]
}

func ConvertBits(x uint32, m int) uint32 {
	v := convertRightBits(x, m) + convertLeftBits(x, m)
	v &= 0xFFFFFF
	return v
}

func convertRightBits(x uint32, m int) uint32 {
	var v uint32 = 0
	for i := 0; i < m; i++ {
		if ((x >> i) & 0x1) == 0x1 {
			v |= (0xAA << (i * 8))
		}
	}
	return v
}

func convertLeftBits(x uint32, m int) uint32 {
	var v uint32 = 0
	x = x >> m
	for i := 0; i < m; i++ {
		if ((x >> i) & 0x1) == 0x1 {
			v |= (0x55 << (i * 8))
		}
	}
	return v
}

func RgbaValues(c color.Color) (r, g, b, a uint8) {
	cr, cg, cb, ca := c.RGBA()
	return uint8(cr >> 8), uint8(cg >> 8), uint8(cb >> 8), uint8(ca >> 8)
}

func HexValue(c color.Color) (h uint32) {
	r, g, b, _ := RgbaValues(c)

	h = 0
	h |= uint32(r) << 16
	h |= uint32(g) << 8
	h |= uint32(b)

	return
}

func CreateColor(h uint32) color.Color {
	v := h & 0xFFFFFF

	r := v >> 16
	r &= 0xFF
	g := v >> 8
	g &= 0xFF
	b := v & 0xFF
	a := 0xFF // force alpha to 1.0

	return color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}
}

func GenerateRandomColor() (h, r, g, b, a uint32, c color.Color) {
	h = rand.Uint32() & 0xFFFFFF
	c = CreateColor(h)
	r, g, b, a = c.RGBA()
	return
}
