package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"image/png"
	"net/http"

	"github.com/fairhive-labs/go-pixelart/internal/filter"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.MaxMultipartMemory = 16 << 20 // 16 MiB
	r.POST("/pixelize", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		fmt.Printf("👉 Source file %q opened\n", file.Filename)

		src, err := file.Open()
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		defer src.Close()
		img, f, err := image.Decode(src)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		fmt.Printf("🤖 Image DECODED - Format is %q\n", f)
		b := img.Bounds()
		fmt.Printf("🖼  Original Dimension = [ %d x %d ]\n", b.Max.X, b.Max.Y)

		fmt.Println("👾 Processing Transformation...")
		ft := filter.NewPixelFilter(100, filter.ShortEdge, filter.CGA4)
		p := ft.Process(&img)
		fmt.Println("✅ Transformation is over")
		switch f {
		case "png":
			buf := bytes.Buffer{}
			err = png.Encode(&buf, p)
			if err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}
			data := base64.StdEncoding.EncodeToString(buf.Bytes())
			c.String(http.StatusOK, data)
			fmt.Printf("🎨 Pixel Art produced\n")
		default:
			c.AbortWithError(http.StatusInternalServerError, errors.New("unsupported picture type"))
			return
		}

	})
	fmt.Println(r.Run())
}
