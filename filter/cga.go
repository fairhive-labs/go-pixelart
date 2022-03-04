package filter

import (
	"image/color"
	"sort"
)

var (
	CGA4  = []uint32{0x0, 0x55FFFF, 0xFF55FF, 0xFFFFFF}
	CGA16 = []uint32{0x0, 0xAA, 0xAA00, 0xAAAA, 0xAA0000, 0xAA00AA, 0xAA5500, 0xAAAAAA,
		0x555555, 0x5555FF, 0x55FF55, 0x55FFFF, 0xFF5555, 0xFF55FF, 0xFFFF55, 0xFFFFFF}
)

func init() {
	sort.Slice(CGA4, func(i, j int) bool { return SortFN(CGA4, i, j) })
	sort.Slice(CGA16, func(i, j int) bool { return SortFN(CGA16, i, j) })
}

func SortFN(s []uint32, i, j int) bool {
	return s[i] < s[j]
}

func CGAColor(c color.Color, f uint32, m ...uint32) color.Color {
	r, g, b, a := c.RGBA()

	var v uint32 = 0
	v |= (r << 16)
	v |= (g << 8)
	v |= b
	v &= 0xffffff

	for i, x := range m {
		if v < x {

			if i == len(m)-1 { // brighter
				v = x
				break
			}

			if (v - m[i-1]) < (x-v)*f { // more contrast
				v = m[i-1]
			} else {
				v = x
			}
			break
		}
	}

	return color.RGBA{uint8(v >> 16), uint8(v >> 8), uint8(v), uint8(a)}
}
