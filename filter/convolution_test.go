package filter

import (
	"errors"
	"testing"
)

func TestNewKernel(t *testing.T) {
	tt := []struct {
		name string
		s    int
		m    Matrix
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
			ErrNilMatrix,
		},
		{
			"empty matrix",
			3,
			[]int{},
			1,
			ErrEmptyMatrix,
		},
		{
			"zero size",
			0,
			[]int{0, 0, 0, 0, 1, 0, 0, 0, 0},
			1,
			ErrKernelSize,
		},
		{
			"too small size",
			2,
			[]int{0, 0, 0, 0, 1, 0, 0, 0, 0},
			1,
			ErrKernelSize,
		},
		{
			"negative size",
			-1,
			[]int{0, 0, 0, 0, 1, 0, 0, 0, 0},
			1,
			ErrKernelSize,
		},
		{
			"even size",
			8,
			[]int{0, 0, 0, 0, 1, 0, 0, 0, 0},
			1,
			ErrKernelSize,
		},
		{
			"malformated matrix",
			3,
			[]int{0, 0, 0, 0, 1, 0, 0, 0, 0, 12345},
			1,
			ErrMalformatedMatrix,
		},
		{
			"factor is zero",
			3,
			[]int{0, 0, 0, 0, 1, 0, 0, 0, 0},
			0,
			ErrIncompatibleFactor,
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
