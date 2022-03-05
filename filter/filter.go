package filter

import (
	"image/color"
	"math/rand"
	"sort"
)

func Transform(c color.Color) color.Color {
	return FastCGA64(c)
}

func convertColor(c color.Color) (r, g, b, a uint8) {
	cr, cg, cb, ca := c.RGBA()
	return uint8(cr >> 8), uint8(cg >> 8), uint8(cb >> 8), uint8(ca >> 8)
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

func GrayColor(c color.Color) color.Color {
	r, g, b, a := convertColor(c)
	v := r/3 + g/3 + b/3
	return color.RGBA{v, v, v, a}
}

func InvertColor(c color.Color) color.Color {
	r, g, b, a := convertColor(c)
	return color.RGBA{0xFF - r, 0xFF - g, 0xFF - b, a}
}

func XRayColor(c color.Color) color.Color {
	return LightGrayColor(InvertColor(c))
}

func DarkGrayColor(c color.Color) color.Color { // get the darkest value in RGBA
	return ConstrastGrayColor(c, 0xFF, func(i, v uint8) bool { return i < v })
}

func LightGrayColor(c color.Color) color.Color { // get the brighest value in RGBA
	return ConstrastGrayColor(c, 0x00, func(i, v uint8) bool { return i > v })
}

type predicate func(uint8, uint8) bool

func ConstrastGrayColor(c color.Color, m uint8, p predicate) color.Color {
	r, g, b, a := convertColor(c)
	s := [3]uint8{r, g, b}
	var v uint8 = m
	for _, i := range s {
		if p(i, v) {
			v = i
		}
	}
	return color.RGBA{v, v, v, a}
}

func DarkContrast(c color.Color) color.Color {
	r, g, b, a := convertColor(c)

	s := []uint8{r, g, b}
	sort.Slice(s, func(i, j int) bool {
		return s[i] <= s[j]
	})

	m := map[uint8]int{r: 0, g: 0, b: 0}
	for i, x := range s {
		m[x] = i
	}

	s[1] -= ((s[1] - s[0]) >> 1)
	s[2] -= ((s[2] - s[0]) >> 2)

	return color.RGBA{s[m[r]], s[m[g]], s[m[b]], a}
}
