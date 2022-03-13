package filter

import (
	"errors"
	"fmt"
	"log"
)

type Matrix []int

type Kernel struct {
	size   int
	matrix Matrix
	factor int
}

const (
	Min int = 3
)

var (
	ErrNilMatrix          = errors.New("kernel matrix cannot be nil")
	ErrEmptyMatrix        = errors.New("kernel matrix cannot be empty")
	ErrKernelSize         = fmt.Errorf("unsupported kernel size, min kernel size = %d", Min)
	ErrMalformatedMatrix  = errors.New("kernel size and matrix length are incompatible")
	ErrIncompatibleFactor = errors.New("kernel factor cannot be 0")

	Identity_3x3 = Kernel{
		3,
		[]int{0, 0, 0, 0, 1, 0, 0, 0, 0},
		1,
	}

	RidgeDetection_3x3 = Kernel{
		3,
		[]int{0, -1, 0, -1, 4, -1, 0, -1, 0},
		1,
	}

	Sharpen_3x3 = Kernel{
		3,
		[]int{0, -1, 0, -1, 5, -1, 0, -1, 0},
		1,
	}
)

func NewKernel(s int, m Matrix, f int) (*Kernel, error) {
	if m == nil {
		return nil, ErrNilMatrix
	}
	if len(m) == 0 {
		return nil, ErrEmptyMatrix
	}

	if s < Min {
		log.Printf("kernel size = %d\n", s)
		return nil, ErrKernelSize
	}
	if s%2 == 0 {
		log.Printf("kernel size = %d, kernel size must be an odd number\n", s)
		return nil, ErrKernelSize
	}
	if s*s != len(m) {
		log.Printf("kernel matrix contains %d elements and shoud contain %d elements\n", len(m), s*s)
		return nil, ErrMalformatedMatrix
	}

	if f == 0 {
		log.Printf("incompatible factor %d", f)
		return nil, ErrIncompatibleFactor
	}

	return &Kernel{s, m, f}, nil
}
