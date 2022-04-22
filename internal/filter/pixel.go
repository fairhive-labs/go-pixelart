package filter

import (
	"fmt"
	"image"
	"image/color"
	"sync"
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
	fmt.Printf("📐 block size = %d pixels\n", blockSize)

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
	colors := make(chan *Block, X*Y)
	var wg sync.WaitGroup
	for x := 0; x < X; x++ {
		for y := 0; y < Y; y++ {
			wg.Add(1)
			go func(x, y int) {
				defer wg.Done()
				if (x%2 == 0 && y%2 == 0) || (x%2 == 1 && y%2 == 1) {
					colors <- &Block{x, y, color.Black}
				} else {
					colors <- &Block{x, y, color.White}
				}
			}(x, y)
		}
	}
	wg.Wait()
	close(colors)
	for b := range colors {
		blockMap[b.x][b.y] = b.c
	}

	xMax, yMax := X*blockSize, Y*blockSize
	fmt.Printf("🖼  New Dimension = [ %d x %d ]\n", xMax, yMax)
	p := image.NewRGBA(image.Rect(0, 0, xMax, yMax)) // create new picture with full blocks
	for x := 0; x < xMax; x++ {
		for y := 0; y < yMax; y++ {
			c := blockMap[x/blockSize][y/blockSize]
			p.Set(x, y, c)
		}
	}
	return p
}
