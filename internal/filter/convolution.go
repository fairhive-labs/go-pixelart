package filter

import (
	"errors"
	"fmt"
	"image"
	"image/color"
)

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

type matrix []int

type kernel struct {
	size   int
	matrix matrix
	factor int
}

type convolutionFilter struct {
	k         *kernel
	pre, post TransformColor
}

func NewConvolutionFilter(k *kernel, pre, post TransformColor) *convolutionFilter {
	return &convolutionFilter{k, pre, post}
}

func (f *convolutionFilter) Process(src image.Image) *image.RGBA {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	p := image.NewRGBA(image.Rect(0, 0, w, h))

	// pre-processing is applied once per pixel here instead of once per kernel
	// tap - a k*k reduction in transformation calls for the same result
	pre := f.pre
	if pre == nil {
		pre = Identity
	}
	src = NewBasicFilter(pre).Process(src)
	at := newSampler(src)

	parallelRange(h, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			i := y * p.Stride
			for x := 0; x < w; x++ {
				c := processConvolution(at, image.Point{}, x, y, w, h, f.k, nil, f.post)
				p.Pix[i], p.Pix[i+1], p.Pix[i+2], p.Pix[i+3] = c.R, c.G, c.B, c.A
				i += 4
			}
		}
	})
	return p
}

func NewKernel(s int, m matrix, f int) (*kernel, error) {
	if m == nil {
		return nil, errNilMatrix
	}
	if len(m) == 0 {
		return nil, errEmptyMatrix
	}

	if s < Min {
		fmt.Printf("kernel size = %d\n", s)
		return nil, errKernelSize
	}
	if s%2 == 0 {
		fmt.Printf("kernel size = %d, kernel size must be an odd number\n", s)
		return nil, errKernelSize
	}
	if s*s != len(m) {
		fmt.Printf("kernel matrix contains %d elements and shoud contain %d elements\n", len(m), s*s)
		return nil, errMalformatedMatrix
	}

	if f == 0 {
		fmt.Printf("incompatible factor %d", f)
		return nil, errIncompatibleFactor
	}

	return &kernel{s, m, f}, nil
}

// processConvolution computes the convolution of the kernel k around (x, y).
// preProcessing may be nil when the caller already applied it to the sampled
// image, which is what convolutionFilter.Process does.
func processConvolution(at sampler, orig image.Point, x, y, xmax, ymax int, k *kernel, preProcessing, postProcessing TransformColor) color.RGBA {
	if postProcessing == nil {
		postProcessing = Identity
	}

	if k == nil {
		return toRGBA(postProcessing(getPixel(preProcessing, at, orig, x, y)))
	}

	s := k.size
	rs := 0 // red accumulator
	gs := 0 // green accumulator
	bs := 0 // blue accumulator

	for i := 0; i < s; i++ {
		for j := 0; j < s; j++ {
			var c color.RGBA
			if i+x-s/2 >= 0 && j+y-s/2 >= 0 && i+x-s/2 < xmax && j+y-s/2 < ymax {
				c = getPixel(preProcessing, at, orig, i+x-s/2, j+y-s/2)
			} else {
				c = getPixel(preProcessing, at, orig, x, y)
			}
			rs += k.matrix[j*s+i] * int(c.R)
			gs += k.matrix[j*s+i] * int(c.G)
			bs += k.matrix[j*s+i] * int(c.B)
		}
	}

	if k.factor != 1 {
		rs /= k.factor
		gs /= k.factor
		bs /= k.factor
	}

	return toRGBA(postProcessing(color.RGBA{uint8(correctValue(rs)), uint8(correctValue(gs)), uint8(correctValue(bs)), 0xFF}))
}

func getPixel(t TransformColor, at sampler, orig image.Point, x, y int) color.RGBA {
	r, g, b, a := at(orig.X+x, orig.Y+y)
	c := color.RGBA{r, g, b, a}
	if t == nil {
		return c
	}
	return toRGBA(t(c))
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
		fmt.Printf("kernel size = %d\n", s)
		return nil, errKernelSize
	}
	if s%2 == 0 {
		fmt.Printf("kernel size = %d, kernel size must be an odd number\n", s)
		return nil, errKernelSize
	}
	n := s * s
	m := make([]int, n)

	for i := range m {
		m[i] = 1
	}

	return NewKernel(s, m, n)
}
