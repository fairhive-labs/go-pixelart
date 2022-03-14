package main

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/fairhive-labs/go-pixelart/filter"
)

func main() {
	src, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal("Cannot open file", os.Args[1], err)
	}
	defer src.Close()
	fmt.Printf("👉 Source file %q opened\n", os.Args[1])

	img, f, err := image.Decode(src)
	if err != nil {
		log.Fatal("Cannot decode: ", err)
	}
	fmt.Printf("🤖 Image DECODED - Format is %q\n", f)

	b := img.Bounds()
	p := image.NewRGBA(image.Rect(0, 0, b.Max.X, b.Max.Y))
	fmt.Printf("🖼  Dimension = [ %d x %d ]\n", b.Max.X, b.Max.Y)

	fmt.Println("👾 Processing Transformation...")
	for x := 0; x < b.Max.X; x++ {
		for y := 0; y < b.Max.Y; y++ {
			// k, _ := filter.Gauss(17)
			k := &filter.RidgeDetection_3x3_hard
			c := filter.ProcessConvolution(k, nil, nil, &img, x, y, b.Max.X, b.Max.Y)
			p.Set(x, y, c)
		}
	}
	fmt.Println("✅ Transformation is over")

	Save(GetFilename(src.Name(), time.Now()), f, p)
}

func GetFilename(f string, t time.Time) string {
	e := filepath.Ext(f)
	n := f[0 : len(f)-len(e)]
	return n + "_" + t.Format("20060102-150405") + e
}

func Save(n, e string, p image.Image) {
	f, err := os.Create(n)
	if err != nil {
		log.Printf("Cannot create file %q", n)
	}
	defer f.Close()

	switch e {
	case "png":
		err = png.Encode(f, p)
	default:
		err = fmt.Errorf("unsupported image format %q", e)
	}

	if err != nil {
		log.Print("Cannot Encode Pixel Art", err)
	}

	fmt.Printf("💾 Pixel Art saved in file %q\n", f.Name())
}
