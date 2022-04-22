package filter

import (
	"image"
	"image/color"
	"sort"

	"github.com/fairhive-labs/go-pixelart/colorutils"
)

type Filter interface {
	//Process image transformation from source src to destination dst
	Process(src, dst *image.Image) error
}

type basicFilter struct {
	transform TransformColor
}

type predicate func(uint8, uint8) bool

type TransformColor func(color.Color) color.Color

func NewBasicFilter(transform TransformColor) *basicFilter {
	return &basicFilter{transform}
}

func (f *basicFilter) Process(src *image.Image, dst *image.RGBA) (err error) {
	b := (*src).Bounds()
	for x := 0; x < b.Max.X; x++ {
		for y := 0; y < b.Max.Y; y++ {
			c := f.transform((*src).At(x, y))
			(*dst).Set(x, y, c)
		}
	}
	return
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
