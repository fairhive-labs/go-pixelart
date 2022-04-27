package filter

import (
	"image"
	"image/color"
	"log"

	"github.com/fairhive-labs/go-pixelart/internal/colorutils"
)

// Edge is a Predicate function
type Edge func(bx, by int) bool

type Block struct {
	x, y int
	c    color.Color
}

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

func (f *pixelFilter) Process(src *image.Image) *image.RGBA {
	b := (*src).Bounds()
	blockSize := f.getBlockSize(b.Max.X, b.Max.Y)
	log.Printf("📐 block size = %d pixels\n", blockSize)

	X := b.Max.X / blockSize
	if b.Max.X%blockSize != 0 {
		X++
	}

	Y := b.Max.Y / blockSize
	if b.Max.Y%blockSize != 0 {
		Y++
	}

	// allocate a block map
	blockMap := make([][]color.Color, X)
	for i := range blockMap {
		blockMap[i] = make([]color.Color, Y)
	}

	// init as checkerboard
	colorsSize := X * Y
	colors := make(chan *Block, colorsSize)
	for x := 0; x < X; x++ {
		for y := 0; y < Y; y++ {
			go func(x, y int) {
				count, ra, ga, ba, aa := 0, 0, 0, 0, 0 // length + RGB averages
				// read colors in original pictures
				for i := x * blockSize; i < (x+1)*blockSize && i < b.Max.X; i++ {
					for j := y * blockSize; j < (y+1)*blockSize && j < b.Max.Y; j++ {
						count++
						c := (*src).At(i, j)
						rc, gc, bc, ac := colorutils.RgbaValues(c)
						ra += int(rc)
						ga += int(gc)
						ba += int(bc)
						aa += int(ac)
					}
				}
				//calcul averages
				ra = ra / count
				ra &= 0xFF
				ga = ga / count
				ga &= 0xFF
				ba = ba / count
				ba &= 0xFF
				aa = aa / count
				aa &= 0xFF
				c := color.RGBA{uint8(ra), uint8(ga), uint8(ba), uint8(aa)}
				colors <- &Block{x, y, c}
			}(x, y)
		}
	}
	// get colors from original picture and transform them
	for i := 0; i < colorsSize; i++ {
		b := <-colors
		blockMap[b.x][b.y] = f.transform(b.c)
	}
	close(colors)

	// create new picture with whole blocks
	xMax, yMax := X*blockSize, Y*blockSize
	log.Printf("🖼  New Dimension = [ %d x %d ]\n", xMax, yMax)
	p := image.NewRGBA(image.Rect(0, 0, xMax, yMax))
	for x := 0; x < xMax; x++ {
		for y := 0; y < yMax; y++ {
			c := blockMap[x/blockSize][y/blockSize] // set the color from the pixel block map
			p.Set(x, y, c)
		}
	}
	return p
}
