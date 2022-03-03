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

	for x := 0; x < b.Max.X; x++ {
		for y := 0; y < b.Max.Y; y++ {
			pic.Set(x, y, img.At(x, y))
		}
	}
	fmt.Println("👾 Transformation : OK")

	pa, err := os.Create("pixel_art." + f)
	if err != nil {
		log.Printf("Cannot create file %q", "pixel_art."+f)
	}
	defer pa.Close()

	err = png.Encode(pa, pic)
	if err != nil {
		log.Print("Cannot Encode Pixel Art", err)
	}
	fmt.Printf("💾 Pixel Art saved in file %q\n", pa.Name())

}

// func save(pic *image.Image, filename string) {
//
// }
