package main

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
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
	pic := image.NewGray(image.Rect(0, 0, b.Max.X, b.Max.Y))

	fmt.Println("👾 Processing Transformation...")
	for x := 0; x < b.Max.X; x++ {
		for y := 0; y < b.Max.Y; y++ {
			pic.Set(x, y, img.At(x, y))
		}
	}
	fmt.Println("✅ Transformation is over")

	save(filename(src.Name()), f, pic)
}

func filename(f string) string {
	return "pixel_art.png"
}

func save(n, f string, pic image.Image) {
	pa, err := os.Create(n)
	if err != nil {
		log.Printf("Cannot create file %q", n)
	}
	defer pa.Close()

	err = png.Encode(pa, pic)
	if err != nil {
		log.Print("Cannot Encode Pixel Art", err)
	}

	fmt.Printf("💾 Pixel Art saved in file %q\n", pa.Name())
}
