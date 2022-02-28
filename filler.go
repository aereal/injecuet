package injecuet

import (
	"fmt"
	"os"
	"strings"

	"cuelang.org/go/cue"
	"github.com/fujiwara/tfstate-lookup/tfstate"
)

// Filler is CUE value filler.
type Filler interface {
	// Name is unique identifier of the Filler.
	Name() string

	// FillValue fills value from the filler's source into given CUE document.
	FillValue(doc Document, key string, field cue.Value) error
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
	fillerNameEnv     = "env"
	fillerNameTFState = "tfstate"
)

type envFillter struct {
	env map[string]string
}

func (f *envFillter) Name() string { return fillerNameEnv }

func (f *envFillter) FillValue(doc Document, key string, field cue.Value) error {
	v, ok := f.env[key]
	if !ok {
		return fmt.Errorf("value not found: %s", key)
	}
	*doc.Value = doc.Value.FillPath(field.Path(), v)
	return doc.Value.Err()
}

func NewTFStateFiller(state *tfstate.TFState) Filler {
	return &tfstateFiller{state: state}
}

type tfstateFiller struct {
	state *tfstate.TFState
}

func (f *tfstateFiller) Name() string { return fillerNameTFState }

func (f *tfstateFiller) FillValue(doc Document, key string, field cue.Value) error {
	obj, err := f.state.Lookup(key)
	if err != nil {
		return fmt.Errorf("tfstate value (%s) not found: %w", key, err)
	}
	value := obj.Value
	if accpetsOnlyInt(field) {
		value, _ = tryDowncastToInt(value)
	}
	*doc.Value = doc.Value.FillPath(field.Path(), value)
	return doc.Value.Err()
}

func accpetsOnlyInt(field cue.Value) bool {
	acceptsInt := field.Unify(field.Context().CompileString("1")).Validate() == nil
	acceptsFloat := field.Unify(field.Context().CompileString("1.0")).Validate() == nil
	return acceptsInt && !acceptsFloat
}

func tryDowncastToInt(x interface{}) (interface{}, bool) {
	switch x := x.(type) {
	case float32:
		return int32(x), true
	case float64:
		return int64(x), true
	default:
		return x, false
	}
}
