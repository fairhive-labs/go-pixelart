package main

import (
	"bytes"
	"embed"
	"encoding/base64"
	"errors"
	"html/template"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/fairhive-labs/go-pixelart/internal/filter"
	"github.com/gin-gonic/gin"
)

type PixelizeForm struct {
	Slices int                   `form:"slices" binding:"required,min=1,max=1000"`
	Edge   string                `form:"edge" binding:"required,oneof=short long"`
	Filter string                `form:"filter" binding:"required,oneof=cga2 cga4 cga16 ega vga identity dark-contrast dark-gray gray invert light-gray xray"`
	File   *multipart.FileHeader `form:"file" binding:"required"`
}

//go:embed templates
var tfs embed.FS

var (
	filters map[string]filter.TransformColor = map[string]filter.TransformColor{
		"cga2":          filter.CGA2,
		"cga4":          filter.CGA4,
		"cga16":         filter.CGA16,
		"ega":           filter.EGA,
		"vga":           filter.VGA,
		"identity":      filter.Identity,
		"dark-contrast": filter.DarkContrast,
		"dark-gray":     filter.DarkGrayColor,
		"gray":          filter.GrayColor,
		"invert":        filter.InvertColor,
		"light-gray":    filter.LightGrayColor,
		"xray":          filter.XRayColor,
	}
	errUnsupportedFilter = errors.New("Unsupported Filter")
)

func main() {
	r := gin.Default()
	t := template.Must(template.New("").Funcs(template.FuncMap{
		"ToUpper": strings.ToUpper,
	}).ParseFS(tfs, "templates/*"))
	r.SetHTMLTemplate(t)
	r.MaxMultipartMemory = 16 << 20 // 16 MiB

	r.Use(cors)
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", filters)
	})
	r.POST("/pixelize", pixelize)
	log.Println(r.Run())
}

func getFilter(f string) (t filter.TransformColor, err error) {
	t, ok := filters[f]
	if !ok {
		return t, errUnsupportedFilter
	}
	return
}

func pixelize(c *gin.Context) {
	var form PixelizeForm
	if err := c.ShouldBind(&form); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	log.Printf("📏 Slices = %d \n", form.Slices)
	log.Printf("👉 Source file %q opened\n", form.File.Filename)

	edge := filter.ShortEdge
	if form.Edge == "long" {
		edge = filter.LongEdge
		log.Println("🏗  Using Long Edge")
	}

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
	if form.Slices > b.Max.X || form.Slices > b.Max.Y {
		log.Println("❌ Transformation aborted")
		c.String(http.StatusBadRequest, "Cannot cut [ %d x %d ] into %d slices", b.Max.X, b.Max.Y, form.Slices)
		return
	}
	fl, err := getFilter(form.Filter)
	if err != nil {
		log.Println("❌ Transformation aborted")
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	ft := filter.NewPixelFilter(form.Slices, edge, fl)
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
	mime := c.DefaultQuery("mime", "html")
	switch mime {
	case "json":
		c.JSON(http.StatusCreated, gin.H{
			"data":     data,
			"encoding": "base64",
		})
	default:
		c.HTML(http.StatusCreated, "pixelart_template.html", gin.H{
			"data": data,
		})
	}
	log.Printf("🎨 Pixel Art produced\n")
}

func cors(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "origin, content-type, accept, authorization")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}
	c.Next()
}
