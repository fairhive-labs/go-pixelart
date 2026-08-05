package filter

import (
	"image"
	"image/color"
	"io"
	"log"
	"os"
	"testing"
)

// TestMain silences the progress lines the filters log, they would otherwise
// be interleaved with the test and benchmark output.
func TestMain(m *testing.M) {
	log.SetOutput(io.Discard)
	os.Exit(m.Run())
}

// benchImage builds a deterministic gradient picture, big enough to make the
// per-pixel cost of the filters visible.
func benchImage(w, h int) image.Image {
	p := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			p.SetRGBA(x, y, color.RGBA{
				uint8(x % 0x100),
				uint8(y % 0x100),
				uint8((x + y) % 0x100),
				0xFF,
			})
		}
	}
	return p
}

func BenchmarkPixelProcess(b *testing.B) {
	src := benchImage(4000, 3000)
	f := NewPixelFilter(200, ShortEdge, EGA)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Process(src)
	}
}

func BenchmarkBasicProcess(b *testing.B) {
	src := benchImage(2000, 1500)
	f := NewBasicFilter(DarkContrast)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Process(src)
	}
}

func BenchmarkConvolutionProcess(b *testing.B) {
	src := benchImage(1000, 750)
	f := NewConvolutionFilter(&Gauss_3x3, Identity, Identity)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Process(src)
	}
}
