package filter

import (
	"image/color"
	"log"
	"sort"
)

var (
	CGA4Table  = []uint32{0x0, 0x55FFFF, 0xFF55FF, 0xFFFFFF}
	CGA16Table = []uint32{0x0, 0xAA, 0xAA00, 0xAAAA, 0xAA0000, 0xAA00AA, 0xAA5500, 0xAAAAAA,
		0x555555, 0x5555FF, 0x55FF55, 0x55FFFF, 0xFF5555, 0xFF55FF, 0xFFFF55, 0xFFFFFF}
	CGA64Table  []uint32
	CGAPalettes map[int]color.Palette
)

func init() {
	CGAPalettes = make(map[int]color.Palette)

	CGA4colors := make([]color.Color, len(CGA4Table))
	for i, e := range CGA4Table {
		c := CreateColor(e)
		CGA4colors[i] = c
	}
	CGAPalettes[4] = CGA4colors

	sort.Slice(CGA16Table, func(i, j int) bool { return sortAsc(CGA16Table, i, j) })

	CGA64Table = initCGA64(64, 3)
	sort.Slice(CGA64Table, func(i, j int) bool { return sortAsc(CGA64Table, i, j) })

}

func CGA(n int, c color.Color) color.Color {
	p, ok := CGAPalettes[n]
	if !ok {
		log.Fatalf("CGA%d not supported\n", n)
	}
	return p.Convert(c)
}

func initCGA64(n, b int) []uint32 {
	s := make([]uint32, n)
	for i := 0; i < n; i++ {
		s[i] = convertBits(uint32(i), b)
	}
	return s
}

func convertBits(x uint32, m int) uint32 {
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

func sortAsc(s []uint32, i, j int) bool {
	return s[i] < s[j]
}

// Deprecated
func CGAColor(c color.Color, t ...uint32) color.Color {
	r, g, b, a := c.RGBA()

	r &= 0xFF
	g &= 0xFF
	b &= 0xFF

	r = (0x4 * r) / 0xFF
	g = (0x4 * g) / 0xFF
	b = (0x4 * b) / 0xFF

	var v uint32 = 0
	v |= (r << 4)
	v |= (g << 2)
	v |= b
	v &= 0xFFFFFF

	v = t[v]

	return color.RGBA{uint8(v >> 16), uint8(v >> 8), uint8(v), uint8(a)}
}
