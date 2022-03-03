package injecuet

import (
	"fmt"
	"io"
	"os"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/parser"
	"github.com/rs/zerolog"
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

// Option is a function to enhance Injector's behavior.
type Option func(i *Injector)

// WithEnvironmentVariables tells Injector to inject environment variables.
func WithEnvironmentVariables(filterEnv func(name string) bool) Option {
	return func(i *Injector) {
		filler := newEnvFillter(filterEnv)
		i.fillers[filler.name()] = filler
	}
}

// WithTFState tells Injector to inject Terraform state.
func WithTFState() Option {
	return func(i *Injector) {
		filler := newTFStateFiller()
		i.fillers[filler.name()] = filler
	}
}

// WithLoggerOutput tells Injector to emit logs into `out`.
func WithLoggerOutput(out io.Writer) Option {
	return func(i *Injector) {
		i.logger = i.logger.Output(out)
	}
}

// WithLogLevel tells Injector to set minimum log level.
func WithLogLevel(level zerolog.Level) Option {
	return func(i *Injector) {
		i.logger = i.logger.Level(level)
	}
}

// WithLogger tells Injector to use given logger.
// This option is for internal use.
func WithLogger(logger zerolog.Logger) Option {
	return func(i *Injector) {
		i.logger = logger
	}
}

// NewInjector creates new Injector.
func NewInjector(options ...Option) *Injector {
	i := &Injector{
		fillers: map[string]filler{},
		logger:  zerolog.New(os.Stderr).With().Caller().Logger(),
	}
	for _, opt := range options {
		opt(i)
	}
	return i
}

// Injector is used for injecting provided values.
// The injection values are given from several constructors.
type Injector struct {
	fillers map[string]filler
	logger  zerolog.Logger
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
			l := i.logger.With().Str("path", value.Path().String()).Logger()
			ret := parseAttribute(value)
			if !ret.valid() {
				l.Info().Err(ret.err).Msg("invalid attribute")
				return
			}
			filler := i.fillers[ret.fillerName]
			if filler == nil {
				l.Warn().Str("kind", ret.fillerName).Msg("not supported kind")
				return
			}
			err = filler.fillValue(document{filename: srcPath, value: &doc}, ret.key, value)
			if err != nil {
				l.Warn().Str("key", ret.key).Err(err).Msg("failed to fill value")
			}
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
