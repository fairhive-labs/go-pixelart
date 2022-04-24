package main

import (
	"bytes"
	"embed"
	"encoding/base64"
	"errors"
	"html/template"
	"image"
	"image/png"
	"log"
	"mime/multipart"
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

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
	r.POST("/pixelize", pixelize)
	log.Println(r.Run())
}

type PixelizeForm struct {
	Slices int                   `form:"slices" binding:"required"`
	File   *multipart.FileHeader `form:"file" binding:"required"`
}

func pixelize(c *gin.Context) {
	var form PixelizeForm
	if err := c.ShouldBind(&form); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	log.Printf("📏 Slices = %d \n", form.Slices)
	log.Printf("👉 Source file %q opened\n", form.File.Filename)

	src, err := form.File.Open()
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
	log.Printf("🤖 Image DECODED - Format is %q\n", f)

	b := img.Bounds()
	log.Printf("🖼  Original Dimension = [ %d x %d ]\n", b.Max.X, b.Max.Y)

	log.Println("👾 Processing Transformation...")
	ft := filter.NewPixelFilter(form.Slices, filter.ShortEdge, filter.CGA64)
	p := ft.Process(&img)
	log.Println("✅ Transformation is over")

	switch f {
	case "png":
		buf := bytes.Buffer{}
		err = png.Encode(&buf, p)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		data := base64.StdEncoding.EncodeToString(buf.Bytes())
		c.HTML(http.StatusCreated, "pixelart_template.html", gin.H{
			"width":  500,
			"height": -1,
			"data":   data,
		})
		log.Printf("🎨 Pixel Art produced\n")
	default:
		c.AbortWithError(http.StatusInternalServerError, errors.New("unsupported picture type"))
	}

}
