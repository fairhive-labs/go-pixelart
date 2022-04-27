package filter

import (
	"image/color"
	"sort"

	"github.com/fairhive-labs/go-pixelart/internal/colorutils"
)

var (
	CGAPalettes map[int]color.Palette
)

func init() {
	CGAPalettes = make(map[int]color.Palette)

	CGAPalettes[2] = color.Palette{color.Black, color.White}

	t4 := []uint32{0x0, 0xFF55FF, 0x55FFFF, 0xFFFFFF}
	CGAPalettes[4] = generatePalette(t4)

	t16 := []uint32{0x0, 0xAA, 0xAA00, 0xAAAA, 0xAA0000, 0xAA00AA, 0xAA5500, 0xAAAAAA,
		0x555555, 0x5555FF, 0x55FF55, 0x55FFFF, 0xFF5555, 0xFF55FF, 0xFFFF55, 0xFFFFFF}
	CGAPalettes[16] = generatePalette(t16)

	t64 := initCGA64Table()
	CGAPalettes[64] = generatePalette(t64)
}

func generatePalette(t []uint32) color.Palette {
	c := make([]color.Color, len(t))
	for i, e := range t {
		c[i] = colorutils.CreateColor(e)
	}
	return c
}

func initCGA64Table() []uint32 {
	s := make([]uint32, 64)
	for i := 0; i < 64; i++ {
		s[i] = colorutils.ConvertBits(uint32(i), 3)
	}
	sort.Slice(s, func(i, j int) bool { return colorutils.SortAsc(s, i, j) })
	return s
}

func EGA(c color.Color) color.Color {
	if colorutils.IsTransparent(c) {
		return color.Transparent
	}
	r, g, b, _ := c.RGBA()
	r &= 0xFF
	g &= 0xFF
	b &= 0xFF

	var m uint32 = 0x3
	if r > (0x100 >> 1) {
		m = 0x4
	}
	r = (m * r) / 0x100
	r &= 0x3

	m = 0x3
	if g > (0x100 >> 1) {
		m = 0x4
	}
	g = (m * g) / 0x100
	g &= 0x3

	m = 0x3
	if b > (0x100 >> 1) {
		m = 0x4
	}
	b = (m * b) / 0x100
	b &= 0x3

	var v uint32 = 0
	v |= (r << 4)
	v |= (g << 2)
	v |= b
	v &= 0x3F

	return CGAPalettes[64][v]
}

func CGA16(c color.Color) color.Color {
	if colorutils.IsTransparent(c) {
		return color.Transparent
	}
	r, g, b, _ := c.RGBA()
	r &= 0xFF
	g &= 0xFF
	b &= 0xFF

	var l uint32 = 0x0
	if (r/3 + g/3 + b/3) > (0x100 >> 1) { // compute the r,g,b average brightness
		l = 0x1
	}

	r = (0x2 * r) / 0x100
	r &= 0x1
	g = (0x2 * g) / 0x100
	g &= 0x1
	b = (0x2 * b) / 0x100
	b &= 0x1

	var v uint32 = 0
	v |= (l << 3)
	v |= (r << 2)
	v |= (g << 1)
	v |= b

	v &= 0xF

	return CGAPalettes[16][v]
}

func CGA4(c color.Color) (n color.Color) {
	if colorutils.IsTransparent(c) {
		return color.Transparent
	}
	h := colorutils.HexValue(c)
	t := CGAPalettes[4]
	var m uint32 = 0x1000000
	i := (h / (m >> 2))
	return t[i]
}

func CGA2(c color.Color) (n color.Color) {
	if colorutils.IsTransparent(c) {
		return color.Transparent
	}
	h := colorutils.HexValue(c)
	t := CGAPalettes[2]
	var m uint32 = 0x1000000
	i := (h / (m >> 1))
	return t[i]
}

func VGA(c color.Color) color.Color {
	if colorutils.IsTransparent(c) {
		return color.Transparent
	}
	r, g, b, _ := c.RGBA()

	r = r >> 2
	r &= 0x3F
	r = (r << 2) | (r >> 4)

	g = g >> 2
	g &= 0x3F
	g = (g << 2) | (g >> 4)

	b = b >> 2
	b &= 0x3F
	b = (b << 2) | (b >> 4)

	return color.RGBA{uint8(r), uint8(g), uint8(b), 0xFF}
}
