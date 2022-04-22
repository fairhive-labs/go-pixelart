package filter

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"log"

	"github.com/fairhive-labs/go-pixelart/colorutils"
)

type matrix []int

type kernel struct {
	size   int
	matrix matrix
	factor int
}

type Convolution struct {
	k       *kernel
	process func(img *image.Image, x, y, xmax, ymax int, k *kernel, preProcessing, postProcessing TransformColor) color.Color
}

const (
	Min int = 3
)

var (
	errNilMatrix          = errors.New("kernel matrix cannot be nil")
	errEmptyMatrix        = errors.New("kernel matrix cannot be empty")
	errKernelSize         = fmt.Errorf("unsupported kernel size, min kernel size = %d", Min)
	errMalformatedMatrix  = errors.New("kernel size and matrix length are incompatible")
	errIncompatibleFactor = errors.New("kernel factor cannot be 0")

	Identity_3x3 = kernel{
		3,
		[]int{0, 0, 0, 0, 1, 0, 0, 0, 0},
		1,
	}

	RidgeDetection_3x3_soft = kernel{
		3,
		[]int{0, -1, 0, -1, 4, -1, 0, -1, 0},
		1,
	}

	RidgeDetection_3x3_hard = kernel{
		3,
		[]int{-1, -1, -1, -1, 8, -1, -1, -1, -1},
		1,
	}

	Sharpen_3x3 = kernel{
		3,
		[]int{0, -1, 0, -1, 5, -1, 0, -1, 0},
		1,
	}

	Gauss_3x3 = kernel{
		3,
		[]int{1, 1, 1, 1, 1, 1, 1, 1, 1},
		9,
	}
)

func NewKernel(s int, m matrix, f int) (*kernel, error) {
	if m == nil {
		return nil, errNilMatrix
	}
	if len(m) == 0 {
		return nil, errEmptyMatrix
	}

	if s < Min {
		log.Printf("kernel size = %d\n", s)
		return nil, errKernelSize
	}
	if s%2 == 0 {
		log.Printf("kernel size = %d, kernel size must be an odd number\n", s)
		return nil, errKernelSize
	}
	if s*s != len(m) {
		log.Printf("kernel matrix contains %d elements and shoud contain %d elements\n", len(m), s*s)
		return nil, errMalformatedMatrix
	}

	if f == 0 {
		log.Printf("incompatible factor %d", f)
		return nil, errIncompatibleFactor
	}

	return &kernel{s, m, f}, nil
}

func ProcessConvolution(img *image.Image, x, y, xmax, ymax int, k *kernel, preProcessing, postProcessing TransformColor) color.Color {
	if postProcessing == nil {
		postProcessing = colorutils.Identity
	}

	if k == nil {
		return postProcessing(getPixel(preProcessing, img, x, y))
	}

	s := k.size
	rs := 0 // red accumulator
	gs := 0 // green accumulator
	bs := 0 // blue accumulator

	for i := 0; i < s; i++ {
		for j := 0; j < s; j++ {
			var c color.Color
			if i+x-s/2 >= 0 && j+y-s/2 >= 0 && i+x-s/2 < xmax && j+y-s/2 < ymax {
				c = getPixel(preProcessing, img, i+x-s/2, j+y-s/2)
			} else {
				c = getPixel(preProcessing, img, x, y)
			}
			r, g, b, _ := colorutils.RgbaValues(c)
			rs += k.matrix[j*s+i] * int(r)
			gs += k.matrix[j*s+i] * int(g)
			bs += k.matrix[j*s+i] * int(b)
		}
	}

	if k.factor != 1 {
		rs /= k.factor
		gs /= k.factor
		bs /= k.factor
	}

	return postProcessing(color.RGBA{uint8(correctValue(rs)), uint8(correctValue(gs)), uint8(correctValue(bs)), 0xFF})
}

func getPixel(t TransformColor, img *image.Image, x, y int) color.Color {
	if t == nil {
		t = colorutils.Identity
	}
	return t((*img).At(x, y))
}

func correctValue(x int) int {
	if x < 0 {
		return 0
	}
	if x > 0xFF {
		return 0xFF
	}
	return x
}

func Gauss(s int) (*kernel, error) {
	if s < Min {
		log.Printf("kernel size = %d\n", s)
		return nil, errKernelSize
	}
	if s%2 == 0 {
		log.Printf("kernel size = %d, kernel size must be an odd number\n", s)
		return nil, errKernelSize
	}
	n := s * s
	m := make([]int, n)

	for i := range m {
		m[i] = 1
	}

	return NewKernel(s, m, n)
}
