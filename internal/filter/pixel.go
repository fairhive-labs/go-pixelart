package filter

import (
	"image"
	"image/color"
	"log"
)

// Edge is a Predicate function
type Edge func(bx, by int) bool

type pixelFilter struct {
	stripes   int
	edge      Edge
	transform TransformColor
}

func NewPixelFilter(stripes int, edge Edge, transform TransformColor) *pixelFilter {
	return &pixelFilter{
		stripes:   stripes,
		edge:      edge,
		transform: transform,
	}
}

func ShortEdge(bx, by int) bool {
	return bx < by
}

func LongEdge(bx, by int) bool {
	return bx > by
}

func (f pixelFilter) getBlockSize(bx, by int) int {
	if f.edge(bx, by) {
		return bx / f.stripes
	}
	return by / f.stripes
}

func (f *pixelFilter) Process(src image.Image) *image.RGBA {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 {
		return image.NewRGBA(image.Rectangle{})
	}

	blockSize := max(f.getBlockSize(w, h), 1) // a block cannot be thinner than a pixel
	log.Printf("📐 block size = %d pixels\n", blockSize)

	X := (w + blockSize - 1) / blockSize
	Y := (h + blockSize - 1) / blockSize

	// average and transform every block, columns of blocks in parallel
	blocks := make([]color.RGBA, X*Y)
	at := newSampler(src)
	parallelRange(X, func(x0, x1 int) {
		for x := x0; x < x1; x++ {
			for y := 0; y < Y; y++ {
				blocks[x*Y+y] = f.blockColor(at, b.Min, x, y, blockSize, w, h)
			}
		}
	})

	// create new picture with whole blocks
	xMax, yMax := X*blockSize, Y*blockSize
	log.Printf("🖼  New Dimension = [ %d x %d ]\n", xMax, yMax)
	p := image.NewRGBA(image.Rect(0, 0, xMax, yMax))
	// paint the first pixel row of each band of blocks, then replicate it down
	// the band: one memmove per row instead of blockSize² Set calls per block
	parallelRange(Y, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			top := y * blockSize * p.Stride
			row := p.Pix[top : top+p.Stride]
			i := 0
			for x := 0; x < X; x++ {
				c := blocks[x*Y+y]
				for range blockSize {
					row[i], row[i+1], row[i+2], row[i+3] = c.R, c.G, c.B, c.A
					i += 4
				}
			}
			for j := 1; j < blockSize; j++ {
				o := top + j*p.Stride
				copy(p.Pix[o:o+p.Stride], row)
			}
		}
	})

	return p
}

// blockColor averages the source pixels covered by block (bx, by) and applies
// the filter transformation to the result.
func (f *pixelFilter) blockColor(at sampler, orig image.Point, bx, by, size, w, h int) color.RGBA {
	x0, y0 := bx*size, by*size
	x1, y1 := min(x0+size, w), min(y0+size, h)

	count, ra, ga, ba, aa := 0, 0, 0, 0, 0 // length + RGBA averages
	for x := x0; x < x1; x++ {
		for y := y0; y < y1; y++ {
			r, g, b, a := at(orig.X+x, orig.Y+y)
			ra += int(r)
			ga += int(g)
			ba += int(b)
			aa += int(a)
			count++
		}
	}

	c := color.RGBA{
		uint8((ra / count) & 0xFF),
		uint8((ga / count) & 0xFF),
		uint8((ba / count) & 0xFF),
		uint8((aa / count) & 0xFF),
	}
	return toRGBA(f.transform(c))
}
