package injecuet

import (
	"fmt"
	"strings"

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
	envFillter := NewEnvFillter(match)
	injector := &Injector{fillers: map[string]Filler{envFillter.Name(): envFillter}}
	return injector
}

// NewInjector creates new Injector.
func NewInjector(fillers ...Filler) *Injector {
	i := &Injector{fillers: map[string]Filler{}}
	for _, f := range fillers {
		i.fillers[f.Name()] = f
	}
	return i
}

// Injector is used for injecting provided values.
// The injection values are given from several constructors.
type Injector struct {
	fillers map[string]Filler
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
			ret, err := parseAttribute(value)
			if err != nil {
				// invalid attribute
				return
			}
			filler := i.fillers[ret.fillerName]
			if filler == nil {
				// not supported filler
				return
			}
			_ = filler.FillValue(&doc, ret.key, value)
		},
	)
	return doc, nil
}

type attributeParseResult struct {
	fillerName string
	key        string
}

func parseAttribute(value cue.Value) (*attributeParseResult, error) {
	if v, _ := parseDeprecatedAttribute(value); v != nil {
		return v, nil
	}
	attr := value.Attribute(attrKey)
	if err := attr.Err(); err != nil {
		return nil, err
	}
	parts := strings.SplitN(attr.Contents(), "=", 2)
	return &attributeParseResult{fillerName: parts[0], key: parts[1]}, nil
}

func parseDeprecatedAttribute(value cue.Value) (*attributeParseResult, error) {
	attr := value.Attribute(deprecatedOldAttrKey)
	if err := attr.Err(); err != nil {
		return nil, err
	}
	return &attributeParseResult{
		fillerName: fillerNameEnv,
		key:        attr.Contents(),
	}, nil
}
