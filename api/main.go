package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"embed"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/fairhive-labs/go-pixelart/internal/filter"
	"github.com/gin-gonic/gin"
)

type PixelizeForm struct {
	Slices int                   `form:"slices" binding:"required,min=1,max=1000"`
	Edge   string                `form:"edge" binding:"required,oneof=short long"`
	Filter string                `form:"filter" binding:"required,oneof=cga2 cga4 cga16 ega vga identity dark-contrast dark-gray gray invert light-gray xray"`
	File   *multipart.FileHeader `form:"file" binding:"required"`
}

//go:embed assets templates
var fs embed.FS

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

func setupRouter() *gin.Engine {
	r := gin.Default()
	t := template.Must(template.New("").Funcs(template.FuncMap{
		"ToUpper": strings.ToUpper,
	}).ParseFS(fs, "templates/*"))
	r.SetHTMLTemplate(t)
	r.MaxMultipartMemory = 16 << 20 // 16 MiB

	r.Use(cors)
	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	r.GET("/favicon.ico", getFavicon)
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", filters)
	})
	r.POST("/pixelize", pixelize)
	return r
}

func main() {
	r := setupRouter()

	var addr string
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	} else {
		addr = ":8080" // default port
	}

	srv := &http.Server{
		Addr:           addr,
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   20 * time.Second,
		IdleTimeout:    time.Minute,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
		s := <-quit
		log.Printf("🚨 Shutdown signal \"%v\" received\n", s)

		log.Printf("🚦 Here we go for a graceful Shutdown...\n")
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			// Error from closing listeners, or context timeout:
			log.Printf("⚠️ HTTP server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()

	log.Printf("✅ Listening and serving HTTP on %s\n", addr)
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("👹 HTTP server ListenAndServe: %v", err)
	}

	<-idleConnsClosed
	log.Printf("😴 Server stopped")

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
			"filter":   form.Filter,
			"length":   len(data),
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
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}
	c.Next()
}

func getFavicon(c *gin.Context) {
	file, err := fs.ReadFile("assets/favicon.ico")
	etag := fmt.Sprintf("%x", md5.Sum(file))
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if match := c.GetHeader("If-None-Match"); match != "" {
		if strings.Contains(match, etag) {
			c.AbortWithStatus(http.StatusNotModified)
			return
		}
	}
	c.Header("Cache-Control", "public, max-age=31536000")
	c.Header("ETag", etag)
	c.Data(
		http.StatusOK,
		"image/x-icon",
		file,
	)
}
