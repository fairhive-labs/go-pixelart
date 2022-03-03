package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"time"
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
		log.Fatal("Cannot decode image", err)
	}
	fmt.Printf("🤖 Image DECODED - Format is %q\n", f)

	b := img.Bounds()
	pic := image.NewRGBA(image.Rect(0, 0, b.Max.X, b.Max.Y))

	fmt.Println("👾 Processing Transformation...")
	for x := 0; x < b.Max.X; x++ {
		for y := 0; y < b.Max.Y; y++ {
			c := img.At(x, y)
			c = Transform(c)
			pic.Set(x, y, c)
		}
	}
	fmt.Println("✅ Transformation is over")

	Save(GetFilename(src.Name(), time.Now()), f, pic)
}

func GetFilename(f string, t time.Time) string {
	e := filepath.Ext(f)
	n := f[0 : len(f)-len(e)]
	return n + "_" + t.Format("20060102-150405") + e
}

func Save(n, e string, pic image.Image) {
	pa, err := os.Create(n)
	if err != nil {
		log.Printf("Cannot create file %q", n)
	}
	defer pa.Close()

	switch e {
	case "png":
		err = png.Encode(pa, pic)
	default:
		err = fmt.Errorf("unsupported image format %q", e)
	}

	if err != nil {
		log.Print("Cannot Encode Pixel Art", err)
	}

	fmt.Printf("💾 Pixel Art saved in file %q\n", pa.Name())
}

func Transform(c color.Color) color.Color {
	return GrayColor(c)
}

func convert(c color.Color) (r, g, b, a uint8) {
	R, G, B, A := c.RGBA()
	return uint8(R), uint8(G), uint8(B), uint8(A)
}

func GrayColor(c color.Color) color.Color {
	return color.GrayModel.Convert(c)
}

func InvertColor(c color.Color) color.Color {
	r, g, b, a := convert(c)
	return color.RGBA{255 - r, 255 - g, 255 - b, a}
}

func DarkGrayColor(c color.Color) color.Color { // get the darkest value in RGBA
	return constrastGrayColor(c, 255, func(i, v uint8) bool { return i < v })
}

func LightGrayColor(c color.Color) color.Color { // get the brighest value in RGBA
	return constrastGrayColor(c, 0, func(i, v uint8) bool { return i > v })
}

type predicate func(uint8, uint8) bool

func constrastGrayColor(c color.Color, m uint8, p predicate) color.Color {
	r, g, b, a := convert(c)
	s := [3]uint8{r, g, b}
	var v uint8 = m
	for _, i := range s {
		if p(i, v) {
			v = i
		}
	}
	return color.RGBA{v, v, v, a}
}

func XRayColor(c color.Color) color.Color {
	return LightGrayColor(InvertColor(c))
}
