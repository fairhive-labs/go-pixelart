package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"time"

	"github.com/fairhive-labs/go-pixelart/internal/filter"
)

func main() {
	src, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Print("Cannot open file", os.Args[1], err)
		os.Exit(1)
	}
	defer src.Close()
	fmt.Printf("👉 Source file %q opened\n", os.Args[1])

	img, f, err := image.Decode(src)
	if err != nil {
		fmt.Print("Cannot decode: ", err)
		os.Exit(1)
	}
	fmt.Printf("🤖 Image DECODED - Format is %q\n", f)

	b := img.Bounds()
	fmt.Printf("🖼  Original Dimension = [ %d x %d ]\n", b.Max.X, b.Max.Y)

	fmt.Println("👾 Processing Transformation...")
	ft := filter.NewPixelFilter(100, filter.ShortEdge, filter.CGA4)
	p := ft.Process(img)
	fmt.Println("✅ Transformation is over")

	if err := save(getFilename(src.Name(), time.Now()), f, p); err != nil {
		fmt.Printf("error saving transformed picture: %v", err)
		os.Exit(1)
	}
}

func getFilename(f string, t time.Time) string {
	e := filepath.Ext(f)
	n := f[0 : len(f)-len(e)]
	return n + "_" + t.Format("20060102-150405") + e
}

func save(n, e string, p image.Image) (err error) {
	f, err := os.Create(n)
	if err != nil {
		fmt.Printf("Cannot create file %q", n)
		return err
	}
	defer f.Close()

	switch e {
	case "png":
		err = png.Encode(f, p)
	default:
		err = fmt.Errorf("unsupported image format %q", e)
	}

	if err != nil {
		fmt.Print("Cannot Encode Pixel Art", err)
	} else {
		fmt.Printf("💾 Pixel Art saved in file %q\n", f.Name())
	}
	return
}
