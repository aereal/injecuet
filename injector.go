package injecuet

import (
	"fmt"
	"os"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/parser"
)

var (
	alwaysReturn = func(_ cue.Value) bool { return true }
	attrKey      = "injectenv"
	matchAll     = func(_ string) bool { return true }
)

// NewEnvironmentInjector returns an new Injector that injects environment variables.
// The injected environment variables determined this function called.
// Modified environment variables are not respected after creating an Injector.
func NewEnvironmentInjector(match func(name string) bool) *Injector {
	if match == nil {
		match = matchAll
	}
	injector := &Injector{injections: map[string]string{}}
	for _, pair := range os.Environ() {
		kv := strings.SplitN(pair, "=", 2)
		if !match(kv[0]) {
			continue
		}
		injector.injections[kv[0]] = kv[1]
	}
	return injector
}

// Injector is used for injecting provided values.
// The injection values are given from several constructors.
type Injector struct {
	injections map[string]string
}

// Inject injects provided injection values to CUE document in srcPath.
func (i *Injector) Inject(srcPath string) (cue.Value, error) {
	f, err := parser.ParseFile(srcPath, nil)
	if err != nil {
		return cue.Value{}, fmt.Errorf("cannot parse file(%s): %w", srcPath, err)
	}
	cc := cuecontext.New()
	doc := cc.BuildFile(f)
	if i.injections == nil {
		return doc, nil
	}
	doc.Walk(
		alwaysReturn,
		func(value cue.Value) {
			attr := value.Attribute(attrKey)
			if err := attr.Err(); err != nil {
				return
			}
			av := attr.Contents()
			filler, ok := i.injections[av]
			if !ok {
				return
			}
			filled := doc.FillPath(value.Path(), filler)
			if err := filled.Err(); err != nil {
				return
			}
			doc = filled
		},
	)
	return doc, nil
}
