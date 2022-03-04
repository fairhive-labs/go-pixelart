package filter

import (
	"testing"
)

func TestHexa(t *testing.T) {
	a := 0x00ff
	b := 0xff
	if a != b {
		t.Errorf("%x == %x, got %v", a, b, a == b)
		t.FailNow()
	}
}
