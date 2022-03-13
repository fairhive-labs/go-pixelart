package filter

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"log"

	"github.com/fairhive-labs/go-pixelart/utils"
)

type Matrix []int

type kernel struct {
	size   int
	matrix Matrix
	factor int
}

const (
	Min int = 3
)

var (
	ErrNilMatrix          = errors.New("kernel matrix cannot be nil")
	ErrEmptyMatrix        = errors.New("kernel matrix cannot be empty")
	ErrKernelSize         = fmt.Errorf("unsupported kernel size, min kernel size = %d", Min)
	ErrMalformatedMatrix  = errors.New("kernel size and matrix length are incompatible")
	ErrIncompatibleFactor = errors.New("kernel factor cannot be 0")

	Identity_3x3 = kernel{
		3,
		[]int{0, 0, 0, 0, 1, 0, 0, 0, 0},
		1,
	}

	RidgeDetection_3x3 = kernel{
		3,
		[]int{0, -1, 0, -1, 4, -1, 0, -1, 0},
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

func NewKernel(s int, m Matrix, f int) (*kernel, error) {
	if m == nil {
		return nil, ErrNilMatrix
	}
	if len(m) == 0 {
		return nil, ErrEmptyMatrix
	}

	if s < Min {
		log.Printf("kernel size = %d\n", s)
		return nil, ErrKernelSize
	}
	if s%2 == 0 {
		log.Printf("kernel size = %d, kernel size must be an odd number\n", s)
		return nil, ErrKernelSize
	}
	if s*s != len(m) {
		log.Printf("kernel matrix contains %d elements and shoud contain %d elements\n", len(m), s*s)
		return nil, ErrMalformatedMatrix
	}

	if f == 0 {
		log.Printf("incompatible factor %d", f)
		return nil, ErrIncompatibleFactor
	}

	return &kernel{s, m, f}, nil
}

func ProcessConvolution(k *kernel, t ColorTransformation, img *image.Image, x, y, xmax, ymax int) color.Color {
	if t == nil {
		t = utils.Identity
	}

	if k == nil {
		return t((*img).At(x, y))
	}

	s := k.size
	rs := 0
	gs := 0
	bs := 0

	for i := 0; i < s; i++ {
		for j := 0; j < s; j++ {
			var c color.Color
			if i+x-s/2 >= 0 && j+y-s/2 >= 0 && i+x-s/2 < xmax && j+y-s/2 < ymax {
				c = (*img).At(i+x-s/2, j+y-s/2)
			} else {
				c = (*img).At(x, y)
			}
			r, g, b, _ := utils.RgbaValues(c)
			rs += k.matrix[j*s+i] * int(r)
			gs += k.matrix[j*s+i] * int(g)
			bs += k.matrix[j*s+i] * int(b)
		}
	}

	if k.factor != 1 {
		rs /= k.factor
		rs = correctValue(rs)
		gs /= k.factor
		gs = correctValue(gs)
		bs /= k.factor
		bs = correctValue(bs)
	}

	return t(color.RGBA{uint8(rs), uint8(gs), uint8(bs), 0xFF})
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
