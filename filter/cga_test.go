package filter

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
)

func TestSortAsc(t *testing.T) {
	a := []uint32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	b := []uint32{10, 8, 9, 7, 6, 5, 0, 4, 3, 2, 1}

	sort.Slice(b, func(i, j int) bool { return sortAsc(b, i, j) })

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
			got := convertBits(tc.x, 3)
			if got != tc.v {
				t.Errorf("convertBits(%d,3) == %x, got %x", tc.x, tc.v, got)
				t.FailNow()
			}
		})
	}
}
