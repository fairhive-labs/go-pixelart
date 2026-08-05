package filter

import (
	"image"
	"image/color"

	"github.com/fairhive-labs/go-pixelart/internal/colorutils"
)

type Filter interface {
	//Process image transformation from source src to destination dst
	Process(src image.Image) *image.RGBA
}

type basicFilter struct {
	transform TransformColor
}

type predicate func(uint8, uint8) bool

type TransformColor func(color.Color) color.Color

func NewBasicFilter(transform TransformColor) *basicFilter {
	return &basicFilter{transform}
}

func (f *basicFilter) Process(src image.Image) *image.RGBA {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	p := image.NewRGBA(image.Rect(0, 0, w, h))
	at := newSampler(src)
	parallelRange(h, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			i := y * p.Stride
			for x := 0; x < w; x++ {
				r, g, bl, a := at(b.Min.X+x, b.Min.Y+y)
				c := toRGBA(f.transform(color.RGBA{r, g, bl, a}))
				p.Pix[i], p.Pix[i+1], p.Pix[i+2], p.Pix[i+3] = c.R, c.G, c.B, c.A
				i += 4
			}
		}
	})
	return p
}

func GrayColor(c color.Color) color.Color {
	r, g, b, a := colorutils.RgbaValues(c)
	v := r/3 + g/3 + b/3
	return color.RGBA{v, v, v, a}
}

func InvertColor(c color.Color) color.Color {
	r, g, b, a := colorutils.RgbaValues(c)
	return color.RGBA{0xFF - r, 0xFF - g, 0xFF - b, a}
}

func XRayColor(c color.Color) color.Color {
	return LightGrayColor(InvertColor(c))
}

func ConstrastGrayColor(c color.Color, m uint8, p predicate) color.Color {
	r, g, b, a := colorutils.RgbaValues(c)
	s := [3]uint8{r, g, b}
	var v uint8 = m
	for _, i := range s {
		if p(i, v) {
			v = i
		}
	}
	return color.RGBA{v, v, v, a}
}

func DarkGrayColor(c color.Color) color.Color { // get the darkest value in RGBA
	return ConstrastGrayColor(c, 0xFF, func(i, v uint8) bool { return i < v })
}

func LightGrayColor(c color.Color) color.Color { // get the brighest value in RGBA
	return ConstrastGrayColor(c, 0x00, func(i, v uint8) bool { return i > v })
}

func DarkContrast(c color.Color) color.Color {
	r, g, b, a := colorutils.RgbaValues(c)

	s := [3]uint8{r, g, b}
	// ascending sort of 3 values, without sort.Slice and its escaping closure
	if s[0] > s[1] {
		s[0], s[1] = s[1], s[0]
	}
	if s[1] > s[2] {
		s[1], s[2] = s[2], s[1]
	}
	if s[0] > s[1] {
		s[0], s[1] = s[1], s[0]
	}

	// rank of each channel in the sorted values, highest index wins on ties
	ir, ig, ib := rank(s, r), rank(s, g), rank(s, b)

	s[1] -= ((s[1] - s[0]) >> 1)
	s[2] -= ((s[2] - s[0]) >> 2)

	return color.RGBA{s[ir], s[ig], s[ib], a}
}

func rank(s [3]uint8, v uint8) int {
	switch {
	case s[2] == v:
		return 2
	case s[1] == v:
		return 1
	default:
		return 0
	}
}

func Identity(c color.Color) color.Color {
	return c
}
