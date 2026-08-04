package main

import (
	"errors"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fairhive-labs/go-pixelart/internal/filter"
)

type options struct {
	slices  int
	edge    string
	filter  string
	out     string
	quality int
}

func usage(fs *flag.FlagSet) func() {
	return func() {
		fmt.Fprintf(fs.Output(), "👾 pixelart - transform a picture into pixel art\n\n")
		fmt.Fprintf(fs.Output(), "usage: %s [options] <picture.png|picture.jpg>\n\noptions:\n", fs.Name())
		fs.PrintDefaults()
		fmt.Fprintf(fs.Output(), "\nfilters: %s\n", strings.Join(filter.Names(), ", "))
	}
}

func parseOptions(fs *flag.FlagSet, args []string) (*options, string, error) {
	o := &options{}
	fs.IntVar(&o.slices, "slices", 100, "number of slices cut along the selected edge")
	fs.StringVar(&o.edge, "edge", "short", "edge used to cut the slices: short or long")
	fs.StringVar(&o.filter, "filter", "cga4", "color filter, see the list below")
	fs.StringVar(&o.out, "out", "", "output file, defaults to <picture>_<timestamp>.<ext>")
	fs.IntVar(&o.quality, "quality", jpeg.DefaultQuality, "JPEG output quality, from 1 to 100")
	fs.Usage = usage(fs)

	if err := fs.Parse(args); err != nil {
		return nil, "", err
	}
	if fs.NArg() != 1 {
		return nil, "", errors.New("exactly one picture is expected")
	}
	if o.slices < 1 {
		return nil, "", fmt.Errorf("slices must be greater than 0, got %d", o.slices)
	}
	if o.edge != "short" && o.edge != "long" {
		return nil, "", fmt.Errorf("edge must be %q or %q, got %q", "short", "long", o.edge)
	}
	if _, err := filter.Get(o.filter); err != nil {
		return nil, "", fmt.Errorf("%w %q, supported filters: %s", err, o.filter, strings.Join(filter.Names(), ", "))
	}
	if o.quality < 1 || o.quality > 100 {
		return nil, "", fmt.Errorf("quality must be between 1 and 100, got %d", o.quality)
	}
	return o, fs.Arg(0), nil
}

func main() {
	fs := flag.NewFlagSet(filepath.Base(os.Args[0]), flag.ContinueOnError)
	o, path, err := parseOptions(fs, os.Args[1:])
	if err != nil {
		if !errors.Is(err, flag.ErrHelp) {
			fmt.Fprintf(os.Stderr, "👹 %v\n\n", err)
			fs.Usage()
		}
		os.Exit(2)
	}

	if err := run(o, path); err != nil {
		fmt.Fprintf(os.Stderr, "👹 %v\n", err)
		os.Exit(1)
	}
}

func run(o *options, path string) error {
	src, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("cannot open %q: %w", path, err)
	}
	defer src.Close()
	fmt.Printf("👉 Source file %q opened\n", path)

	img, f, err := image.Decode(src)
	if err != nil {
		return fmt.Errorf("cannot decode %q, only PNG and JPEG are supported: %w", path, err)
	}
	fmt.Printf("🤖 Image DECODED - Format is %q\n", f)

	b := img.Bounds()
	fmt.Printf("🖼  Original Dimension = [ %d x %d ]\n", b.Dx(), b.Dy())
	if o.slices > b.Dx() || o.slices > b.Dy() {
		return fmt.Errorf("cannot cut [ %d x %d ] into %d slices", b.Dx(), b.Dy(), o.slices)
	}

	edge := filter.ShortEdge
	if o.edge == "long" {
		edge = filter.LongEdge
	}
	t, err := filter.Get(o.filter)
	if err != nil {
		return err
	}

	fmt.Println("👾 Processing Transformation...")
	p := filter.NewPixelFilter(o.slices, edge, t).Process(img)
	fmt.Println("✅ Transformation is over")

	out := o.out
	if out == "" {
		out = getFilename(path, time.Now())
	}
	return save(out, f, o.quality, p)
}

func getFilename(f string, t time.Time) string {
	e := filepath.Ext(f)
	n := f[0 : len(f)-len(e)]
	return n + "_" + t.Format("20060102-150405") + e
}

func save(n, e string, quality int, p image.Image) error {
	f, err := os.Create(n)
	if err != nil {
		return fmt.Errorf("cannot create file %q: %w", n, err)
	}
	defer f.Close()

	switch e {
	case "png":
		err = png.Encode(f, p)
	case "jpeg":
		err = jpeg.Encode(f, p, &jpeg.Options{Quality: quality})
	default:
		err = fmt.Errorf("unsupported image format %q", e)
	}
	if err != nil {
		return fmt.Errorf("cannot encode the pixel art: %w", err)
	}

	fmt.Printf("💾 Pixel Art saved in file %q\n", f.Name())
	return nil
}
