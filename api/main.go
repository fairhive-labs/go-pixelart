package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"embed"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"html/template"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync"
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

const (
	defaultMaxUploadBytes   int64 = 10 << 20   // 10 MiB
	defaultMaxPixels        int64 = 40_000_000 // 40 MP, ~160 MB once decoded
	multipartMemoryFraction int64 = 4          // part of an upload kept in memory
	shutdownGrace                 = 10 * time.Second

	csp = "default-src 'self'; " +
		"img-src 'self' data:; " +
		"style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; " +
		"script-src 'self' 'unsafe-inline'; " +
		"form-action 'self'; frame-ancestors 'none'; base-uri 'none'"
)

var (
	filters              = filter.ByName
	errUnsupportedFilter = filter.ErrUnsupportedFilter

	// limits, overridable through the environment
	maxUploadBytes = envInt64("MAX_UPLOAD_BYTES", defaultMaxUploadBytes)
	maxPixels      = envInt64("MAX_PIXELS", defaultMaxPixels)

	// jobs bounds the number of transformations running at the same time: each
	// one is CPU bound and holds a decoded picture plus its copy in memory
	jobs = make(chan struct{}, max(int(envInt64("MAX_CONCURRENT_JOBS", int64(runtime.GOMAXPROCS(0)))), 1))

	// the favicon is embedded, its ETag never changes while running
	favicon = sync.OnceValues(func() ([]byte, string) {
		f, err := fs.ReadFile("assets/favicon.ico")
		if err != nil {
			log.Printf("👹 cannot read the embedded favicon: %v", err)
			return nil, ""
		}
		sum := sha256.Sum256(f)
		return f, `"` + hex.EncodeToString(sum[:16]) + `"`
	})
)

func envInt64(name string, def int64) int64 {
	v, err := strconv.ParseInt(os.Getenv(name), 10, 64)
	if err != nil || v <= 0 {
		return def
	}
	return v
}

func setupRouter() *gin.Engine {
	r := gin.Default()
	t := template.Must(template.New("").Funcs(template.FuncMap{
		"ToUpper": strings.ToUpper,
	}).ParseFS(fs, "templates/*"))
	r.SetHTMLTemplate(t)
	r.MaxMultipartMemory = max(maxUploadBytes/multipartMemoryFraction, 1<<20)

	r.Use(secureHeaders, cors)
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
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := setupRouter()

	var addr string
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	} else {
		addr = ":8080" // default port
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       60 * time.Second, // an upload weighs up to maxUploadBytes
		WriteTimeout:      120 * time.Second,
		IdleTimeout:       time.Minute,
		MaxHeaderBytes:    1 << 20, // 1 MB
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
		s := <-quit
		log.Printf("🚨 Shutdown signal \"%v\" received\n", s)

		log.Printf("🚦 Here we go for a graceful Shutdown...\n")
		ctx, cancel := context.WithTimeout(context.Background(), shutdownGrace)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			// Error from closing listeners, or context timeout:
			log.Printf("⚠️ HTTP server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()

	log.Printf("✅ Listening and serving HTTP on %s\n", addr)
	log.Printf("🔒 Limits: %d bytes per upload, %d pixels per picture, %d concurrent jobs\n",
		maxUploadBytes, maxPixels, cap(jobs))
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("👹 HTTP server ListenAndServe: %v", err)
	}

	<-idleConnsClosed
	log.Printf("😴 Server stopped")

}

func getFilter(f string) (filter.TransformColor, error) {
	return filter.Get(f)
}

func pixelize(c *gin.Context) {
	// bound the request body before anything reads it: without this a client
	// can stream an unlimited amount of data into memory and temporary files
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadBytes)

	var form PixelizeForm
	if err := c.ShouldBind(&form); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			c.String(http.StatusRequestEntityTooLarge, "picture is too large, %d bytes max", maxUploadBytes)
			return
		}
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	log.Printf("📏 Slices = %d \n", form.Slices)
	log.Printf("👉 Source file %q opened\n", sanitize(form.File.Filename))

	if form.File.Size > maxUploadBytes {
		c.String(http.StatusRequestEntityTooLarge, "picture is too large, %d bytes max", maxUploadBytes)
		return
	}

	edge := filter.ShortEdge
	if form.Edge == "long" {
		edge = filter.LongEdge
		log.Println("🏗  Using Long Edge")
	}

	fl, err := getFilter(form.Filter)
	if err != nil {
		log.Println("❌ Transformation aborted")
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	src, err := form.File.Open()
	if err != nil {
		log.Printf("👹 cannot open the uploaded picture: %v", err)
		c.String(http.StatusInternalServerError, "cannot read the uploaded picture")
		return
	}
	defer src.Close()

	// read the declared dimensions before decoding: a small file can expand
	// into gigabytes of pixels once decoded
	cfg, f, err := image.DecodeConfig(src)
	if err != nil {
		c.String(http.StatusUnsupportedMediaType, "unsupported picture format, only PNG and JPEG are supported")
		return
	}
	if f != "png" && f != "jpeg" {
		c.String(http.StatusUnsupportedMediaType, "unsupported picture format %q, only PNG and JPEG are supported", f)
		return
	}
	if px := int64(cfg.Width) * int64(cfg.Height); px > maxPixels {
		log.Printf("❌ Transformation aborted, [ %d x %d ] declared\n", cfg.Width, cfg.Height)
		c.String(http.StatusRequestEntityTooLarge,
			"picture is too large, [ %d x %d ] declared, %d pixels max", cfg.Width, cfg.Height, maxPixels)
		return
	}
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		log.Printf("👹 cannot rewind the uploaded picture: %v", err)
		c.String(http.StatusInternalServerError, "cannot read the uploaded picture")
		return
	}

	// bound the number of transformations in flight
	select {
	case jobs <- struct{}{}:
		defer func() { <-jobs }()
	default:
		log.Println("🛑 Too many transformations in flight")
		c.Header("Retry-After", "5")
		c.String(http.StatusServiceUnavailable, "too many pictures being processed, please retry")
		return
	}

	img, f, err := image.Decode(src)
	if err != nil {
		log.Printf("👹 cannot decode the uploaded picture: %v", err)
		c.String(http.StatusUnsupportedMediaType, "cannot decode the uploaded picture")
		return
	}
	log.Printf("🤖 Image DECODED - Format is %q\n", f)

	b := img.Bounds()
	log.Printf("🖼  Original Dimension = [ %d x %d ]\n", b.Dx(), b.Dy())

	log.Println("👾 Processing Transformation...")
	if form.Slices > b.Dx() || form.Slices > b.Dy() {
		log.Println("❌ Transformation aborted")
		c.String(http.StatusBadRequest, "Cannot cut [ %d x %d ] into %d slices", b.Dx(), b.Dy(), form.Slices)
		return
	}
	ft := filter.NewPixelFilter(form.Slices, edge, fl)
	p := ft.Process(img)
	log.Println("✅ Transformation is over")

	buf := bytes.Buffer{}
	switch f {
	case "png":
		err = png.Encode(&buf, p)
	case "jpeg":
		err = jpeg.Encode(&buf, p, nil) //default quality is 75%
	default:
		c.String(http.StatusUnsupportedMediaType, "unsupported picture format")
		return
	}
	if err != nil {
		log.Printf("👹 cannot encode the pixel art: %v", err)
		c.String(http.StatusInternalServerError, "cannot encode the pixel art")
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

// sanitize keeps user controlled values from injecting line breaks and control
// characters into the logs.
func sanitize(s string) string {
	return strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7F {
			return -1
		}
		return r
	}, s)
}

func cors(c *gin.Context) {
	// public API, no cookie and no credential involved: "*" is fine, but
	// Access-Control-Allow-Credentials must not come with it - browsers reject
	// the combination and it would open credentialed cross origin calls
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "origin, content-type, accept")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

	if c.Request.Method == http.MethodOptions {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}
	c.Next()
}

func secureHeaders(c *gin.Context) {
	h := c.Writer.Header()
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("X-Frame-Options", "DENY")
	h.Set("Referrer-Policy", "no-referrer")
	h.Set("Content-Security-Policy", csp)
	c.Next()
}

func getFavicon(c *gin.Context) {
	file, etag := favicon()
	if file == nil {
		c.String(http.StatusInternalServerError, "cannot read the favicon")
		return
	}
	if match := c.GetHeader("If-None-Match"); match != "" {
		for _, m := range strings.Split(match, ",") {
			if m := strings.TrimSpace(m); m == etag || m == "*" {
				c.AbortWithStatus(http.StatusNotModified)
				return
			}
		}
	}
	c.Header("Cache-Control", "public, max-age=31536000, immutable")
	c.Header("ETag", etag)
	c.Data(
		http.StatusOK,
		"image/x-icon",
		file,
	)
}
