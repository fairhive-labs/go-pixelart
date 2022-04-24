package main

import (
	"bytes"
	"embed"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"image"
	"image/png"
	"net/http"

	"github.com/fairhive-labs/go-pixelart/internal/filter"
	"github.com/gin-gonic/gin"
)

//go:embed templates
var tfs embed.FS

func main() {
	r := gin.Default()
	t := template.Must(template.ParseFS(tfs, "templates/*"))
	r.SetHTMLTemplate(t)
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
		ft := filter.NewPixelFilter(200, filter.ShortEdge, filter.CGA64)
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
			c.HTML(http.StatusOK, "pixelart_template.html", gin.H{
				"width":  500,
				"height": -1,
				"data":   data,
			})
			fmt.Printf("🎨 Pixel Art produced\n")
		default:
			c.AbortWithError(http.StatusInternalServerError, errors.New("unsupported picture type"))
		}

	})
	fmt.Println(r.Run())
}
