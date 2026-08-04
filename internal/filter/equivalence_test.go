package filter

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"sort"

	"github.com/fairhive-labs/go-pixelart/internal/colorutils"
	"testing"
)

// The filters were rewritten for speed. Their output must not change: the
// reference implementations below are the original, straightforward versions
// kept as an oracle.

func referencePixelProcess(f *pixelFilter, src image.Image) *image.RGBA {
	b := src.Bounds()
	blockSize := f.getBlockSize(b.Max.X, b.Max.Y)

	X := b.Max.X / blockSize
	if b.Max.X%blockSize != 0 {
		X++
	}
	Y := b.Max.Y / blockSize
	if b.Max.Y%blockSize != 0 {
		Y++
	}

	blockMap := make([][]color.Color, X)
	for i := range blockMap {
		blockMap[i] = make([]color.Color, Y)
	}

	for x := 0; x < X; x++ {
		for y := 0; y < Y; y++ {
			count, ra, ga, ba, aa := 0, 0, 0, 0, 0
			for i := x * blockSize; i < (x+1)*blockSize && i < b.Max.X; i++ {
				for j := y * blockSize; j < (y+1)*blockSize && j < b.Max.Y; j++ {
					count++
					rc, gc, bc, ac := src.At(i, j).RGBA()
					ra += int(uint8(rc >> 8))
					ga += int(uint8(gc >> 8))
					ba += int(uint8(bc >> 8))
					aa += int(uint8(ac >> 8))
				}
			}
			ra = (ra / count) & 0xFF
			ga = (ga / count) & 0xFF
			ba = (ba / count) & 0xFF
			aa = (aa / count) & 0xFF
			blockMap[x][y] = f.transform(color.RGBA{uint8(ra), uint8(ga), uint8(ba), uint8(aa)})
		}
	}

	xMax, yMax := X*blockSize, Y*blockSize
	p := image.NewRGBA(image.Rect(0, 0, xMax, yMax))
	for x := 0; x < xMax; x++ {
		for y := 0; y < yMax; y++ {
			p.Set(x, y, blockMap[x/blockSize][y/blockSize])
		}
	}
	return p
}

func referenceBasicProcess(t TransformColor, src image.Image) *image.RGBA {
	b := src.Bounds()
	p := image.NewRGBA(image.Rect(0, 0, b.Max.X, b.Max.Y))
	for x := 0; x < b.Max.X; x++ {
		for y := 0; y < b.Max.Y; y++ {
			p.Set(x, y, t(src.At(x, y)))
		}
	}
	return p
}

func referenceDarkContrast(c color.Color) color.Color {
	cr, cg, cb, ca := c.RGBA()
	r, g, b, a := uint8(cr>>8), uint8(cg>>8), uint8(cb>>8), uint8(ca>>8)

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

func loadImage(t *testing.T, path string) image.Image {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("cannot open %q: %v", path, err)
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		t.Fatalf("cannot decode %q: %v", path, err)
	}
	return img
}

func assertSameImage(t *testing.T, name string, got, want *image.RGBA) {
	t.Helper()
	if got.Rect != want.Rect {
		t.Fatalf("%s: incorrect bounds, got %v, want %v", name, got.Rect, want.Rect)
	}
	for y := got.Rect.Min.Y; y < got.Rect.Max.Y; y++ {
		for x := got.Rect.Min.X; x < got.Rect.Max.X; x++ {
			if got.RGBAAt(x, y) != want.RGBAAt(x, y) {
				t.Fatalf("%s: pixel (%d,%d) differs, got %v, want %v",
					name, x, y, got.RGBAAt(x, y), want.RGBAAt(x, y))
			}
		}
	}
}

func TestPixelProcessMatchesReference(t *testing.T) {
	allTransforms := map[string]TransformColor{
		"cga2": CGA2, "cga4": CGA4, "cga16": CGA16, "ega": EGA, "vga": VGA,
		"identity": Identity, "dark-contrast": DarkContrast, "dark-gray": DarkGrayColor,
		"gray": GrayColor, "invert": InvertColor, "light-gray": LightGrayColor, "xray": XRayColor,
	}
	// every filter on small pictures, a couple of filters on a real one
	cases := []struct {
		name       string
		img        image.Image
		slices     []int
		transforms map[string]TransformColor
	}{
		{"img_test.png", loadImage(t, "../../api/img_test.png"), []int{1, 7, 33, 100}, allTransforms},
		{"synthetic-rgba", benchImage(320, 240), []int{1, 7, 33, 100}, allTransforms},
		{"synthetic-nrgba", nrgbaImage(200, 150), []int{1, 7, 33, 100}, allTransforms},
		{"wfvr.png", loadImage(t, "../../pictures/wfvr.png"), []int{100},
			map[string]TransformColor{"ega": EGA, "dark-contrast": DarkContrast}},
	}
	edges := map[string]Edge{"short": ShortEdge, "long": LongEdge}

	for _, tc := range cases {
		for _, slices := range tc.slices {
			for ename, edge := range edges {
				for tname, transform := range tc.transforms {
					f := NewPixelFilter(slices, edge, transform)
					if f.getBlockSize(tc.img.Bounds().Dx(), tc.img.Bounds().Dy()) < 1 {
						continue // out of the reference implementation's range
					}
					name := fmt.Sprintf("%s/%d/%s/%s", tc.name, slices, ename, tname)
					t.Run(name, func(t *testing.T) {
						assertSameImage(t, name, f.Process(tc.img), referencePixelProcess(f, tc.img))
					})
				}
			}
		}
	}
}

func TestBasicProcessMatchesReference(t *testing.T) {
	img := loadImage(t, "../../api/img_test.png")
	for name, transform := range map[string]TransformColor{
		"gray": GrayColor, "invert": InvertColor, "xray": XRayColor,
		"dark-contrast": DarkContrast, "dark-gray": DarkGrayColor, "light-gray": LightGrayColor,
		"ega": EGA, "cga16": CGA16,
	} {
		t.Run(name, func(t *testing.T) {
			f := NewBasicFilter(transform)
			assertSameImage(t, name, f.Process(img), referenceBasicProcess(transform, img))
		})
	}
}

// referenceConvolutionProcess is the original sequential implementation:
// pre-processing applied at every kernel tap, one p.Set per pixel.
func referenceConvolutionProcess(k *kernel, pre, post TransformColor, src image.Image) *image.RGBA {
	b := src.Bounds()
	if post == nil {
		post = Identity
	}
	if pre == nil {
		pre = Identity
	}
	pixel := func(x, y int) (uint8, uint8, uint8) {
		r, g, bl, _ := colorutils.RgbaValues(pre(src.At(x, y)))
		return r, g, bl
	}

	p := image.NewRGBA(image.Rect(0, 0, b.Max.X, b.Max.Y))
	for x := 0; x < b.Max.X; x++ {
		for y := 0; y < b.Max.Y; y++ {
			s := k.size
			rs, gs, bs := 0, 0, 0
			for i := 0; i < s; i++ {
				for j := 0; j < s; j++ {
					px, py := i+x-s/2, j+y-s/2
					if px < 0 || py < 0 || px >= b.Max.X || py >= b.Max.Y {
						px, py = x, y
					}
					r, g, bl := pixel(px, py)
					rs += k.matrix[j*s+i] * int(r)
					gs += k.matrix[j*s+i] * int(g)
					bs += k.matrix[j*s+i] * int(bl)
				}
			}
			if k.factor != 1 {
				rs /= k.factor
				gs /= k.factor
				bs /= k.factor
			}
			p.Set(x, y, post(color.RGBA{
				uint8(correctValue(rs)), uint8(correctValue(gs)), uint8(correctValue(bs)), 0xFF,
			}))
		}
	}
	return p
}

func TestConvolutionProcessMatchesReference(t *testing.T) {
	img := loadImage(t, "../../api/img_test.png")

	kernels := map[string]*kernel{
		"identity": &Identity_3x3,
		"sharpen":  &Sharpen_3x3,
		"gauss":    &Gauss_3x3,
		"ridge":    &RidgeDetection_3x3_hard,
	}
	processings := map[string][2]TransformColor{
		"identity/identity": {Identity, Identity},
		"gray/invert":       {GrayColor, InvertColor},
		"nil/nil":           {nil, nil},
	}

	for kname, k := range kernels {
		for pname, p := range processings {
			name := kname + "/" + pname
			t.Run(name, func(t *testing.T) {
				f := NewConvolutionFilter(k, p[0], p[1])
				assertSameImage(t, name, f.Process(img), referenceConvolutionProcess(k, p[0], p[1], img))
			})
		}
	}
}

func TestDarkContrastMatchesReference(t *testing.T) {
	for r := 0; r < 0x100; r += 7 {
		for g := 0; g < 0x100; g += 11 {
			for b := 0; b < 0x100; b += 13 {
				for _, a := range []uint8{0x00, 0x7F, 0xFF} {
					c := color.RGBA{uint8(r), uint8(g), uint8(b), a}
					if got, want := DarkContrast(c), referenceDarkContrast(c); got != want {
						t.Fatalf("DarkContrast(%v) = %v, want %v", c, got, want)
					}
				}
			}
		}
	}
}

// TestSamplerMatchesAt checks the fast paths of newSampler against the generic
// image.Image.At + colorutils.RgbaValues path they replace.
func TestSamplerMatchesAt(t *testing.T) {
	images := map[string]image.Image{
		"rgba":     benchImage(64, 48),
		"nrgba":    nrgbaImage(64, 48),
		"gray":     grayImage(64, 48),
		"ycbcr":    ycbcrImage(64, 48),
		"subimage": benchImage(64, 48).(*image.RGBA).SubImage(image.Rect(10, 8, 40, 30)),
		"paletted": palettedImage(64, 48),
	}
	for name, img := range images {
		t.Run(name, func(t *testing.T) {
			at := newSampler(img)
			b := img.Bounds()
			for y := b.Min.Y; y < b.Max.Y; y++ {
				for x := b.Min.X; x < b.Max.X; x++ {
					r1, g1, b1, a1 := at(x, y)
					cr, cg, cb, ca := img.At(x, y).RGBA()
					r2, g2, b2, a2 := uint8(cr>>8), uint8(cg>>8), uint8(cb>>8), uint8(ca>>8)
					if r1 != r2 || g1 != g2 || b1 != b2 || a1 != a2 {
						t.Fatalf("pixel (%d,%d): sampler = %v, At = %v",
							x, y, []uint8{r1, g1, b1, a1}, []uint8{r2, g2, b2, a2})
					}
				}
			}
		})
	}
}

// TestProcessTinyImage covers pictures smaller than the requested number of
// slices: the block size used to fall to 0 and panic.
func TestProcessTinyImage(t *testing.T) {
	src := benchImage(4, 3)
	p := NewPixelFilter(100, ShortEdge, CGA4).Process(src)
	if p.Bounds().Dx() != 4 || p.Bounds().Dy() != 3 {
		t.Errorf("incorrect bounds, got %v, want 4x3", p.Bounds())
	}
}

func nrgbaImage(w, h int) *image.NRGBA {
	p := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			p.SetNRGBA(x, y, color.NRGBA{
				uint8(x % 0x100), uint8(y % 0x100), uint8((x * y) % 0x100), uint8((x + y) % 0x100),
			})
		}
	}
	return p
}

func grayImage(w, h int) *image.Gray {
	p := image.NewGray(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			p.SetGray(x, y, color.Gray{uint8((x + y) % 0x100)})
		}
	}
	return p
}

func ycbcrImage(w, h int) *image.YCbCr {
	p := image.NewYCbCr(image.Rect(0, 0, w, h), image.YCbCrSubsampleRatio420)
	for i := range p.Y {
		p.Y[i] = uint8(i % 0x100)
	}
	for i := range p.Cb {
		p.Cb[i] = uint8((i * 3) % 0x100)
		p.Cr[i] = uint8((i * 7) % 0x100)
	}
	return p
}

func palettedImage(w, h int) *image.Paletted {
	p := image.NewPaletted(image.Rect(0, 0, w, h), CGAPalettes[16])
	for i := range p.Pix {
		p.Pix[i] = uint8(i % 16)
	}
	return p
}
