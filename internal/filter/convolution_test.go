package filter

import (
	"errors"
	"testing"
)

func TestNewKernel(t *testing.T) {
	tt := []struct {
		name string
		s    int
		m    matrix
		f    int
		err  error
	}{
		{
			"identity matrix",
			3,
			[]int{0, 0, 0, 0, 1, 0, 0, 0, 0},
			1,
			nil,
		},
		{
			"nil matrix",
			3,
			nil,
			1,
			errNilMatrix,
		},
		{
			"empty matrix",
			3,
			[]int{},
			1,
			errEmptyMatrix,
		},
		{
			"zero size",
			0,
			[]int{0, 0, 0, 0, 1, 0, 0, 0, 0},
			1,
			errKernelSize,
		},
		{
			"too small size",
			2,
			[]int{0, 0, 0, 0, 1, 0, 0, 0, 0},
			1,
			errKernelSize,
		},
		{
			"negative size",
			-1,
			[]int{0, 0, 0, 0, 1, 0, 0, 0, 0},
			1,
			errKernelSize,
		},
		{
			"even size",
			8,
			[]int{0, 0, 0, 0, 1, 0, 0, 0, 0},
			1,
			errKernelSize,
		},
		{
			"malformated matrix",
			3,
			[]int{0, 0, 0, 0, 1, 0, 0, 0, 0, 12345},
			1,
			errMalformatedMatrix,
		},
		{
			"factor is zero",
			3,
			[]int{0, 0, 0, 0, 1, 0, 0, 0, 0},
			0,
			errIncompatibleFactor,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewKernel(tc.s, tc.m, tc.f)
			if !errors.Is(err, tc.err) {
				t.Errorf("got error : %v != %v", err, tc.err)
				t.FailNow()
			}
		})
	}

}

func TestCorrectValue(t *testing.T) {
	tt := []struct {
		name   string
		value  int
		result int
	}{
		{"less than 0", -1, 0},
		{"equal 0", 0, 0},
		{"greather than 0xFF", 1024, 0xFF},
		{"equal 0xFF", 0xFF, 0xFF},
		{"128", 128, 128},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got := correctValue(tc.value)
			if got != tc.result {
				t.Errorf("correctValue(%d) = %d , got = %d", tc.value, tc.result, got)
				t.FailNow()
			}
		})
	}
}
