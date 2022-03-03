package injecuet

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue"
	"github.com/fujiwara/tfstate-lookup/tfstate"
)

type filler interface {
	name() string
	fillValue(doc document, key string, field cue.Value) error
}

func newEnvFillter(match func(name string) bool) filler {
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

func (f *envFillter) name() string { return fillerNameEnv }

func (f *envFillter) fillValue(doc document, key string, field cue.Value) error {
	v, ok := f.env[key]
	if !ok {
		return fmt.Errorf("value not found: %s", key)
	}
	*doc.value = doc.value.FillPath(field.Path(), v)
	return doc.value.Err()
}

func newTFStateFiller() filler {
	return &tfstateFiller{}
}

type tfstateFiller struct {
	state *tfstate.TFState
}

func (f *tfstateFiller) name() string { return fillerNameTFState }

func (f *tfstateFiller) fillValue(doc document, key string, field cue.Value) error {
	if f.state == nil {
		attrs := doc.value.Attributes(cue.DeclAttr)
		var initialized bool
		for _, attr := range attrs {
			url := getTFStateURL(attr)
			if url == "" {
				continue
			}
			resolved, err := resolveURL(filepath.Dir(doc.filename), url)
			if err != nil {
				return fmt.Errorf("cannot resolve tfstate URL: %w", err)
			}
			f.state, err = tfstate.ReadURL(resolved)
			if err != nil {
				return fmt.Errorf("cannot read tfstate(%s): %w", url, err)
			}
			initialized = true
			break
		}
		if !initialized {
			return fmt.Errorf("tfstate-lookup is not initialized")
		}
	}
	obj, err := f.state.Lookup(key)
	if err != nil {
		return fmt.Errorf("tfstate value (%s) not found: %w", key, err)
	}
	value := obj.Value
	if accpetsOnlyInt(field) {
		value, _ = tryDowncastToInt(value)
	}
	*doc.value = doc.value.FillPath(field.Path(), value)
	return doc.value.Err()
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

func getTFStateURL(attr cue.Attribute) string {
	var ok bool
	for i := 0; i < attr.NumArgs(); i++ {
		k, v := attr.Arg(i)
		if k == fillerNameTFState && v == "" {
			ok = true
			continue
		}
		if ok && k == "stateURL" {
			return v
		}
	}
	return ""
}

func resolveURL(base string, rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	switch u.Scheme {
	case "file", "":
		return filepath.Clean(filepath.Join(base, u.Path)), nil
	default:
		return "", fmt.Errorf("invalid scheme")
	}
}
