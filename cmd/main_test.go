package main

import (
	"flag"
	"image"
	_ "image/jpeg"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGetFilename(t *testing.T) {
	f := "jsie.png"

	date := "Thu Mar 3 15:04:05 2022"
	ti, err := time.Parse("Mon Jan 2 15:04:05 2006", date)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	got := getFilename(f, ti)
	if got != "jsie_20220303-150405.png" {
		t.Errorf("GetFilename(f, ti) = %s; want jsie_20220303-150405.png", got)
		t.FailNow()
	}
}

func newFlagSet() *flag.FlagSet {
	fs := flag.NewFlagSet("pixelart", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs
}

func TestParseOptions(t *testing.T) {
	tt := []struct {
		name string
		args []string
		ok   bool
	}{
		{"defaults", []string{"picture.png"}, true},
		{"all options", []string{"-slices", "80", "-edge", "long", "-filter", "cga16", "-out", "out.png", "picture.png"}, true},
		{"no picture", []string{}, false},
		{"several pictures", []string{"a.png", "b.png"}, false},
		{"unknown filter", []string{"-filter", "sepia", "picture.png"}, false},
		{"unknown edge", []string{"-edge", "diagonal", "picture.png"}, false},
		{"zero slice", []string{"-slices", "0", "picture.png"}, false},
		{"negative slices", []string{"-slices", "-3", "picture.png"}, false},
		{"quality too high", []string{"-quality", "101", "picture.png"}, false},
		{"quality too low", []string{"-quality", "0", "picture.png"}, false},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := parseOptions(newFlagSet(), tc.args)
			if tc.ok && err != nil {
				t.Errorf("incorrect error, got %v, want nil", err)
			}
			if !tc.ok && err == nil {
				t.Error("an error was expected, got nil")
			}
		})
	}
}

func TestParseOptionsValues(t *testing.T) {
	o, path, err := parseOptions(newFlagSet(),
		[]string{"-slices", "80", "-edge", "long", "-filter", "cga16", "-out", "out.png", "picture.png"})
	if err != nil {
		t.Fatalf("incorrect error, got %v, want nil", err)
	}
	if path != "picture.png" {
		t.Errorf("incorrect picture, got %q, want %q", path, "picture.png")
	}
	if o.slices != 80 || o.edge != "long" || o.filter != "cga16" || o.out != "out.png" {
		t.Errorf("incorrect options: %+v", o)
	}
}

func TestRun(t *testing.T) {
	out := filepath.Join(t.TempDir(), "pixelart.png")
	o, path, err := parseOptions(newFlagSet(),
		[]string{"-slices", "20", "-edge", "short", "-filter", "ega", "-out", out, "../api/img_test.png"})
	if err != nil {
		t.Fatalf("incorrect error parsing the options, got %v, want nil", err)
	}

	if err := run(o, path); err != nil {
		t.Fatalf("incorrect error, got %v, want nil", err)
	}

	f, err := os.Open(out)
	if err != nil {
		t.Fatalf("cannot open the produced file: %v", err)
	}
	defer f.Close()
	img, format, err := image.Decode(f)
	if err != nil {
		t.Fatalf("cannot decode the produced file: %v", err)
	}
	if format != "png" {
		t.Errorf("incorrect format, got %q, want %q", format, "png")
	}
	if img.Bounds().Dx() == 0 || img.Bounds().Dy() == 0 {
		t.Errorf("incorrect dimensions, got %v", img.Bounds())
	}
}

func TestRunErrors(t *testing.T) {
	tt := []struct {
		name string
		args []string
		want string
	}{
		{"missing file", []string{"does-not-exist.png"}, "cannot open"},
		{"not a picture", []string{"main.go"}, "cannot decode"},
		{"too many slices", []string{"-slices", "1000", "../api/img_test.png"}, "cannot cut"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			o, path, err := parseOptions(newFlagSet(), tc.args)
			if err != nil {
				t.Fatalf("incorrect error parsing the options, got %v, want nil", err)
			}
			o.out = filepath.Join(t.TempDir(), "out.png")

			if err := run(o, path); err == nil {
				t.Fatal("an error was expected, got nil")
			} else if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("incorrect error, got %q, want it to contain %q", err, tc.want)
			}
		})
	}
}
