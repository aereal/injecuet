package injecuet

import (
	"fmt"
	"os"
	"strings"

	"cuelang.org/go/cue"
)

// Filler is CUE value filler.
type Filler interface {
	// Name is unique identifier of the Filler.
	Name() string

	// FillValue fills value from the filler's source into given CUE document.
	FillValue(doc *cue.Value, key string, field cue.Value) error
}

// NewEnvFiller returns an new Filler that fills environment variables.
// The filled values are determined this function called.
// Modified environment variables are not respected after creating an Filler.
func NewEnvFillter(match func(name string) bool) Filler {
	if match == nil {
		match = matchAll
	}
	filler := &envFillter{env: map[string]string{}}
	for _, pair := range os.Environ() {
		kv := strings.SplitN(pair, "=", 2)
		if !match(kv[0]) {
			continue
		}
		filler.env[kv[0]] = kv[1]
	}
	return filler
}

const (
	fillerNameEnv = "env"
)

type envFillter struct {
	env map[string]string
}

func (f *envFillter) Name() string { return fillerNameEnv }

func (f *envFillter) FillValue(doc *cue.Value, key string, field cue.Value) error {
	v, ok := f.env[key]
	if !ok {
		return fmt.Errorf("value not found: %s", key)
	}
	*doc = doc.FillPath(field.Path(), v)
	return doc.Err()
}
