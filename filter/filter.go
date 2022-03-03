package filter

import "image/color"

func Transform(c color.Color) color.Color {
	return DarkGrayColor(c)
}

func convert(c color.Color) (r, g, b, a uint8) {
	R, G, B, A := c.RGBA()
	return uint8(R), uint8(G), uint8(B), uint8(A)
}

func GrayColor(c color.Color) color.Color {
	r, g, b, a := convert(c)
	v := r/3 + g/3 + b/3
	return color.RGBA{v, v, v, a}
}

func InvertColor(c color.Color) color.Color {
	r, g, b, a := convert(c)
	return color.RGBA{255 - r, 255 - g, 255 - b, a}
}

func DarkGrayColor(c color.Color) color.Color { // get the darkest value in RGBA
	return constrastGrayColor(c, 255, func(i, v uint8) bool { return i < v })
}

func LightGrayColor(c color.Color) color.Color { // get the brighest value in RGBA
	return constrastGrayColor(c, 0, func(i, v uint8) bool { return i > v })
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
