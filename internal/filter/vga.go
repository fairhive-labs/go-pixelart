package filter

import "image/color"

func VGA(c color.Color) color.Color {
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
