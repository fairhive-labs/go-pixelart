package main

import (
	"bytes"
	"embed"
	"encoding/base64"
	"html/template"
	"image"
	"image/jpeg"
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
	Slices    int                   `form:"slices" binding:"required,min=1,max=1000"`
	Width     int                   `form:"width"`
	ShortEdge string                `form:"edge" binding:"required"`
	File      *multipart.FileHeader `form:"file" binding:"required"`
}

func pixelize(c *gin.Context) {
	var form PixelizeForm
	if err := c.ShouldBind(&form); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	log.Printf("📏 Slices = %d \n", form.Slices)
	log.Printf("👉 Source file %q opened\n", form.File.Filename)

	src, err := form.File.Open()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer src.Close()

	img, f, err := image.Decode(src)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	log.Printf("🤖 Image DECODED - Format is %q\n", f)

	b := img.Bounds()
	log.Printf("🖼  Original Dimension = [ %d x %d ]\n", b.Max.X, b.Max.Y)

	log.Println("👾 Processing Transformation...")
	ft := filter.NewPixelFilter(form.Slices, filter.ShortEdge, filter.CGA64)
	p := ft.Process(&img)
	log.Println("✅ Transformation is over")

	buf := bytes.Buffer{}
	switch f {
	case "png":
		err = png.Encode(&buf, p)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
	case "jpeg":
		err = jpeg.Encode(&buf, p, nil) //default quality is 75%
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
	default:
		c.String(http.StatusInternalServerError, "unsupported picture format")
		return
	}
	data := base64.StdEncoding.EncodeToString(buf.Bytes())
	c.HTML(http.StatusCreated, "pixelart_template.html", gin.H{
		"width": form.Width,
		"data":  data,
	})
	log.Printf("🎨 Pixel Art produced\n")
}
