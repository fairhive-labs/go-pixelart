package filter

import (
	"image/color"
)

func Transform(c color.Color) color.Color {
	return CGAColor(c, 2, CGA4...)
}

func convert(c color.Color) (r, g, b, a uint8) {
	R, G, B, A := c.RGBA()
	return uint8(R >> 8), uint8(G >> 8), uint8(B >> 8), uint8(A >> 8)
}

func GrayColor(c color.Color) color.Color {
	r, g, b, a := convert(c)
	v := r/3 + g/3 + b/3
	return color.RGBA{v, v, v, a}
}

func InvertColor(c color.Color) color.Color {
	r, g, b, a := convert(c)
	return color.RGBA{0xFF - r, 0xFF - g, 0xFF - b, a}
}

func DarkGrayColor(c color.Color) color.Color { // get the darkest value in RGBA
	return constrastGrayColor(c, 0xFF, func(i, v uint8) bool { return i < v })
}

func LightGrayColor(c color.Color) color.Color { // get the brighest value in RGBA
	return constrastGrayColor(c, 0x00, func(i, v uint8) bool { return i > v })
}

type predicate func(uint8, uint8) bool

func constrastGrayColor(c color.Color, m uint8, p predicate) color.Color {
	r, g, b, a := convert(c)
	s := [3]uint8{r, g, b}
	var v uint8 = m
	for _, i := range s {
		if p(i, v) {
			v = i
		}
	}
	return color.RGBA{v, v, v, a}
}

func XRayColor(c color.Color) color.Color {
	return LightGrayColor(InvertColor(c))
}
