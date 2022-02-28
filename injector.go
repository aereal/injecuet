package injecuet

import (
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/parser"
)

var (
	matchAll = func(_ string) bool { return true }

	attrKey              = "inject"
	deprecatedOldAttrKey = "injectenv"
)

// NewEnvironmentInjector returns an new Injector that injects environment variables.
//
// Deprecated: use NewInjector
func NewEnvironmentInjector(match func(name string) bool) *Injector {
	envFillter := newEnvFillter(match)
	injector := &Injector{fillers: map[string]filler{envFillter.name(): envFillter}}
	return injector
}

type Option func(i *Injector)

func WithEnvironmentVariables(filterEnv func(name string) bool) Option {
	return func(i *Injector) {
		filler := newEnvFillter(filterEnv)
		i.fillers[filler.name()] = filler
	}
}

func WithTFState() Option {
	return func(i *Injector) {
		filler := newTFStateFiller()
		i.fillers[filler.name()] = filler
	}
}

// NewInjector creates new Injector.
func NewInjector(options ...Option) *Injector {
	i := &Injector{fillers: map[string]filler{}}
	for _, opt := range options {
		opt(i)
	}
	return i
}

// Injector is used for injecting provided values.
// The injection values are given from several constructors.
type Injector struct {
	fillers map[string]filler
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
	walk(
		doc,
		func(value cue.Value) {
			ret := parseAttribute(value)
			if !ret.valid() {
				// invalid attribute
				return
			}
			filler := i.fillers[ret.fillerName]
			if filler == nil {
				// not supported filler
				return
			}
			_ = filler.fillValue(document{filename: srcPath, value: &doc}, ret.key, value)
		},
	)
	return doc, nil
}

type attributeParseResult struct {
	fillerName string
	key        string
	err        error
}

func (r *attributeParseResult) valid() bool {
	return r.err == nil
}

func parseAttribute(value cue.Value) *attributeParseResult {
	if v := parseDeprecatedAttribute(value); v.valid() {
		return v
	}
	attr := value.Attribute(attrKey)
	if err := attr.Err(); err != nil {
		return &attributeParseResult{err: err}
	}
	ret := &attributeParseResult{}
	for i := 0; i < attr.NumArgs(); i++ {
		key, value := attr.Arg(i)
		if value == "" {
			ret.fillerName = key
			continue
		}
		switch key {
		case "name":
			ret.key = value
		}
	}
	return ret
}

func parseDeprecatedAttribute(value cue.Value) *attributeParseResult {
	attr := value.Attribute(deprecatedOldAttrKey)
	if err := attr.Err(); err != nil {
		return &attributeParseResult{err: err}
	}
	return &attributeParseResult{
		fillerName: fillerNameEnv,
		key:        attr.Contents(),
	}
}

type document struct {
	filename string
	value    *cue.Value
}
