package filter

import (
	"image"
	"image/color"
	"runtime"
	"sync"

	"github.com/fairhive-labs/go-pixelart/internal/colorutils"
)

// parallelRange splits [0, n) into contiguous ranges and hands each one to a
// worker, using at most GOMAXPROCS goroutines. Workers must only touch their
// own range so no synchronization is needed beyond the final barrier.
func parallelRange(n int, fn func(start, end int)) {
	if n <= 0 {
		return
	}
	workers := min(runtime.GOMAXPROCS(0), n)
	if workers <= 1 {
		fn(0, n)
		return
	}

	chunk := (n + workers - 1) / workers
	var wg sync.WaitGroup
	for start := 0; start < n; start += chunk {
		wg.Add(1)
		go func(s, e int) {
			defer wg.Done()
			fn(s, e)
		}(start, min(start+chunk, n))
	}
	wg.Wait()
}

// sampler reads a source pixel as premultiplied 8 bit RGBA. image.Image.At
// boxes every pixel into a color.Color interface, which costs an allocation
// per pixel; the concrete image types below are read directly instead. Values
// are identical to colorutils.RgbaValues(src.At(x, y)) in every case.
type sampler func(x, y int) (r, g, b, a uint8)

func newSampler(src image.Image) sampler {
	switch s := src.(type) {
	case *image.RGBA:
		return func(x, y int) (uint8, uint8, uint8, uint8) {
			i := s.PixOffset(x, y)
			p := s.Pix[i : i+4 : i+4]
			return p[0], p[1], p[2], p[3]
		}
	case *image.NRGBA:
		return func(x, y int) (uint8, uint8, uint8, uint8) {
			i := s.PixOffset(x, y)
			p := s.Pix[i : i+4 : i+4]
			if p[3] == 0xFF { // opaque: premultiplication is a no-op
				return p[0], p[1], p[2], p[3]
			}
			return premultiply(p[0], p[1], p[2], p[3])
		}
	case *image.Gray:
		return func(x, y int) (uint8, uint8, uint8, uint8) {
			v := s.Pix[s.PixOffset(x, y)]
			return v, v, v, 0xFF
		}
	case *image.YCbCr:
		return func(x, y int) (uint8, uint8, uint8, uint8) {
			r, g, b := color.YCbCrToRGB(
				s.Y[s.YOffset(x, y)],
				s.Cb[s.COffset(x, y)],
				s.Cr[s.COffset(x, y)],
			)
			return r, g, b, 0xFF
		}
	default:
		return func(x, y int) (uint8, uint8, uint8, uint8) {
			return colorutils.RgbaValues(src.At(x, y))
		}
	}
}

// premultiply applies the same alpha premultiplication as color.NRGBA.RGBA().
func premultiply(r, g, b, a uint8) (uint8, uint8, uint8, uint8) {
	cr, cg, cb, ca := uint32(r), uint32(g), uint32(b), uint32(a)
	cr |= cr << 8
	cr = cr * ca / 0xFF
	cg |= cg << 8
	cg = cg * ca / 0xFF
	cb |= cb << 8
	cb = cb * ca / 0xFF
	return uint8(cr >> 8), uint8(cg >> 8), uint8(cb >> 8), a
}

// toRGBA converts a transformed color to color.RGBA, short-cutting the common
// case where the transformation already returned one.
func toRGBA(c color.Color) color.RGBA {
	if v, ok := c.(color.RGBA); ok {
		return v
	}
	r, g, b, a := c.RGBA()
	return color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
}
