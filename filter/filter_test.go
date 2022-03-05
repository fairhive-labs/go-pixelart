package filter

import (
	"fmt"
	"image/color"
	"reflect"
	"testing"
)

func TestHex(t *testing.T) {

	for i := 0; i < 10; i++ {
		h, _, _, _, _, c := GenerateRandomColor()
		got := hexValue(c)
		t.Run(fmt.Sprintf("%.6x", h), func(t *testing.T) {
			if got != h {
				t.Errorf("color %v should be %x , got %x", c, h, got)
				t.FailNow()
			}
		})
	}

}

func TestRgba(t *testing.T) {
	var r, g, b, a uint8 = 0x3E, 0x61, 0x43, 0xFF
	c := color.RGBA{r, g, b, a}

	cr, cg, cb, ca := rgbaValues(c)

	exp := []uint8{r, g, b, a}
	got := []uint8{cr, cg, cb, ca}

	if !reflect.DeepEqual(exp, got) {
		t.Errorf("colors are different, exp = %v, got = %v", exp, got)
		t.FailNow()
	}
}

func TestGenerateRandomColor(t *testing.T) {

	n := 20

	for i := 0; i < n; i++ {
		h, r, g, b, a, c := GenerateRandomColor()
		got := color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}

		t.Run(fmt.Sprintf("%.6X", h), func(t *testing.T) {
			if got != c {
				t.Errorf("%v == color.RGBA(%v), got %v", c, []uint32{r, g, b, a}, got)
				t.FailNow()
			}
		})
	}

}

func TestCreateColor(t *testing.T) {

	tt := []struct {
		h uint32
		c color.Color
	}{
		{0xFFFFFF, color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}},
		{0x0, color.RGBA{0, 0, 0, 0xFF}},
		{0x55FFFF, color.RGBA{0x55, 0xFF, 0xFF, 0xFF}},
		{0xFF55FF, color.RGBA{0xFF, 0x55, 0xFF, 0xFF}},
		{0xAA00FF, color.RGBA{0xAA, 0x00, 0xFF, 0xFF}},
		{0x123456, color.RGBA{0x12, 0x34, 0x56, 0xFF}},
		{0xABCDEF, color.RGBA{0xAB, 0xCD, 0xEF, 0xFF}},
		{0x6F6C64, color.RGBA{0x6F, 0x6C, 0x64, 0xFF}},
	}

	for _, tc := range tt {
		t.Run(fmt.Sprintf("%.6X", tc.h), func(t *testing.T) {
			got := CreateColor(tc.h)
			if got != tc.c {
				t.Errorf("%.6X color = %v color, got %v\n", tc.h, tc.c, got)
				t.FailNow()
			}
		})
	}
}
