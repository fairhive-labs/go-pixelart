package filter

import (
	"testing"
)

func TestNewPixelFilter(t *testing.T) {
	s := 10
	e := ShortEdge
	tr := InvertColor
	f := NewPixelFilter(s, e, tr)
	if f == nil {
		t.Errorf("cannot create new pixel filter")
		t.FailNow()
	}
}

func TestEdge(t *testing.T) {
	tt := []struct {
		name   string
		edge   Edge
		bx, by int
		result bool
	}{
		{"ShortEdge(10,20)", ShortEdge, 10, 20, true},
		{"ShortEdge(20,10)", ShortEdge, 20, 10, false},
		{"ShortEdge(10,10)", ShortEdge, 10, 10, false},
		{"LongEdge(10,20)", LongEdge, 10, 20, false},
		{"LongEdge(20,10)", LongEdge, 20, 10, true},
		{"LongEdge(10,10)", LongEdge, 10, 10, false},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			v := tc.edge(tc.bx, tc.by)
			if v != tc.result {
				t.Errorf("incorrect result, got %v, want %v", v, tc.result)
			}
		})
	}
}

func TestGetBlockSize(t *testing.T) {
	tt := []struct {
		name            string
		edge            Edge
		bx, by, stripes int
		result          int
	}{
		{"ShortEdge 10 20 2", ShortEdge, 10, 20, 2, 5},
		{"ShortEdge 20 10 2", ShortEdge, 20, 10, 2, 5},
		{"ShortEdge 10 10 2", ShortEdge, 10, 10, 2, 5},
		{"LongEdge 10 20 2", LongEdge, 10, 20, 2, 10},
		{"LongEdge 20 10 2", LongEdge, 20, 10, 2, 10},
		{"LongEdge 10 10 2", LongEdge, 10, 10, 2, 5},
		{"ShortEdge 10 20 3", ShortEdge, 10, 20, 3, 3},
		{"LongEdge 10 20 3", LongEdge, 10, 20, 3, 6},
	}

	for _, tc := range tt {
		f := pixelFilter{tc.stripes, tc.edge, InvertColor}
		if v := f.getBlockSize(tc.bx, tc.by); v != tc.result {
			t.Errorf("incorrect blocksize, got %d, want %d", v, tc.result)
			t.FailNow()
		}
	}
}
