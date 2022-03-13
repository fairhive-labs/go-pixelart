package utils

import (
	"fmt"
	"image/color"
	"reflect"
	"sort"
	"testing"
)

func TestSortAsc(t *testing.T) {
	a := []uint32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	b := []uint32{10, 8, 9, 7, 6, 5, 0, 4, 3, 2, 1}

	sort.Slice(b, func(i, j int) bool { return SortAsc(b, i, j) })

	if !reflect.DeepEqual(a, b) {
		t.Errorf("%v != %v", a, b)
		t.FailNow()
	}
}

func TestConvertRightBits(t *testing.T) {

	tt := []struct {
		x uint32
		v uint32
	}{
		{0, 0},
		{1, 0xAA},
		{2, 0xAA00},
		{3, 0xAAAA},
		{4, 0xAA0000},
		{5, 0xAA00AA},
		{6, 0xAAAA00},
		{7, 0xAAAAAA},
	}

	for _, tc := range tt {
		t.Run(fmt.Sprintf("%d %.6b", tc.x, tc.x), func(t *testing.T) {
			got := convertRightBits(tc.x, 3)
			if got != tc.v {
				t.Errorf("convertRightBits(%d,3) == %x, got %x", tc.x, tc.v, got)
				t.FailNow()
			}
		})
	}

}

func TestConvertLeftBits(t *testing.T) {

	tt := []struct {
		x uint32
		v uint32
	}{
		{0, 0},
		{8, 0x55},
		{16, 0x5500},
		{24, 0x5555},
		{32, 0x550000},
		{40, 0x550055},
		{48, 0x555500},
		{56, 0x555555},
	}

	for _, tc := range tt {
		t.Run(fmt.Sprintf("%d %.6b", tc.x, tc.x), func(t *testing.T) {
			got := convertLeftBits(tc.x, 3)
			if got != tc.v {
				t.Errorf("convertLeftBits(%d,3) == %x, got %x", tc.x, tc.v, got)
				t.FailNow()
			}
		})
	}

}

func TestConvertBits(t *testing.T) {
	tt := []struct {
		x uint32
		v uint32
	}{
		{0, 0},
		{1, 0xAA},
		{2, 0xAA00},
		{3, 0xAAAA},
		{4, 0xAA0000},
		{5, 0xAA00AA},
		{6, 0xAAAA00},
		{7, 0xAAAAAA},
		{8, 0x55},
		{10, 0xAA55},
		{16, 0x5500},
		{19, 0xFFAA},
		{24, 0x5555},
		{32, 0x550000},
		{40, 0x550055},
		{42, 0x55AA55},
		{47, 0xFFAAFF},
		{48, 0x555500},
		{52, 0xFF5500},
		{56, 0x555555},
	}

	for _, tc := range tt {
		t.Run(fmt.Sprintf("%d %.6b", tc.x, tc.x), func(t *testing.T) {
			got := ConvertBits(tc.x, 3)
			if got != tc.v {
				t.Errorf("convertBits(%d,3) == %x, got %x", tc.x, tc.v, got)
				t.FailNow()
			}
		})
	}
}

func TestHex(t *testing.T) {

	for i := 0; i < 10; i++ {
		h, _, _, _, _, c := GenerateRandomColor()
		got := HexValue(c)
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

	cr, cg, cb, ca := RgbaValues(c)

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
