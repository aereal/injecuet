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
	matchAll     = func(_ string) bool { return true }

	attrKey              = "inject"
	deprecatedOldAttrKey = "injectenv"
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

func walk(v cue.Value, f func(v cue.Value)) {
	switch v.Kind() {
	case cue.StructKind:
		st, _ := v.Struct()
		fields := st.Fields(cue.All())
		for fields.Next() {
			fv := fields.Value()
			walk(fv, f) // TODO: use goto?
		}
	case cue.ListKind:
		list, _ := v.List()
		for list.Next() {
			lv := list.Value()
			walk(lv, f) // TODO: use goto?
		}
	default:
		f(v)
	}
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
	walk(
		doc,
		func(value cue.Value) {
			key := parseKey(value)
			if key == "" {
				return
			}
			filler, ok := i.injections[key]
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

func parseKey(value cue.Value) string {
	attr := value.Attribute(attrKey)
	if err := attr.Err(); err != nil {
		return parseDeprecatedAttributeFormatKey(value)
	}
	parts := strings.SplitN(attr.Contents(), "=", 2)
	return parts[1]
}

func parseDeprecatedAttributeFormatKey(value cue.Value) string {
	attr := value.Attribute(deprecatedOldAttrKey)
	if err := attr.Err(); err != nil {
		return ""
	}
	return attr.Contents()
}
