package filter

import (
	"errors"
	"sort"
)

// ErrUnsupportedFilter is returned when a filter name is not part of ByName.
var ErrUnsupportedFilter = errors.New("unsupported filter")

// ByName maps the public filter names - the ones accepted by the API form and
// by the CLI - to their color transformation.
var ByName = map[string]TransformColor{
	"cga2":          CGA2,
	"cga4":          CGA4,
	"cga16":         CGA16,
	"ega":           EGA,
	"vga":           VGA,
	"identity":      Identity,
	"dark-contrast": DarkContrast,
	"dark-gray":     DarkGrayColor,
	"gray":          GrayColor,
	"invert":        InvertColor,
	"light-gray":    LightGrayColor,
	"xray":          XRayColor,
}

// Get returns the transformation registered under name.
func Get(name string) (TransformColor, error) {
	t, ok := ByName[name]
	if !ok {
		return nil, ErrUnsupportedFilter
	}
	return t, nil
}

// Names lists the supported filter names, sorted.
func Names() []string {
	n := make([]string, 0, len(ByName))
	for k := range ByName {
		n = append(n, k)
	}
	sort.Strings(n)
	return n
}
