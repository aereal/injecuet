package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"

	"cuelang.org/go/cue/format"
	"github.com/aereal/injecuet"
)

var version string

func init() {
	if version == "" {
		version = "latest"
	}
}

type App struct {
	errOut io.Writer
}

func (a *App) Run(argv []string) int {
	if a.errOut == nil {
		a.errOut = os.Stderr
	}

	fs := flag.NewFlagSet(argv[0], flag.ContinueOnError)
	var (
		outPath     string
		showVersion bool
		pattern     string
	)
	fs.StringVar(&outPath, "output", "", "output file path. default is stdout")
	fs.BoolVar(&showVersion, "version", false, "show version")
	fs.StringVar(&pattern, "pattern", "", "regular expression of environment variables' names to consume; the pattern must be valid as Go's regexp")
	fs.SetOutput(a.errOut)
	switch err := fs.Parse(argv[1:]); err {
	case flag.ErrHelp:
		return 0
	case nil:
		// skip
	default:
		return 1
	}
	if showVersion {
		fmt.Fprintln(a.errOut, version)
		return 0
	}

	if fs.NArg() != 1 {
		fmt.Fprintln(a.errOut, "input file must be given")
		return 1
	}

	opts := runOptions{
		srcPath: fs.Arg(0),
		outPath: outPath,
		pattern: pattern,
	}
	if err := a.runMain(opts); err != nil {
		fmt.Fprintf(a.errOut, "%v\n", err)
		return 1
	}

	return 0
}

func (a *App) runMain(opts runOptions) error {
	match, err := opts.buildMatchFunction()
	if err != nil {
		return err
	}
	out, close, err := opts.openOutput()
	if err != nil {
		return err
	}
	defer close()

	injector := injecuet.NewInjector(injecuet.WithEnvironmentVariables(match), injecuet.WithTFState())
	v, err := injector.Inject(opts.srcPath)
	if err != nil {
		return fmt.Errorf("failed to inject values to file %s: %w", opts.srcPath, err)
	}
	formatted, err := format.Node(v.Syntax())
	if err != nil {
		return fmt.Errorf("failed to format file %s: %w", opts.srcPath, err)
	}
	_, _ = out.Write(formatted)
	_, _ = out.Write([]byte{'\n'})
	return nil
}

type runOptions struct {
	srcPath string
	outPath string
	pattern string
}

func (o runOptions) buildMatchFunction() (func(string) bool, error) {
	var match func(string) bool
	if o.pattern == "" {
		return match, nil
	}
	re, err := regexp.Compile(o.pattern)
	if err != nil {
		return nil, fmt.Errorf("cannot parse pattern: %w", err)
	}
	return re.MatchString, nil
}

func (o runOptions) openOutput() (io.Writer, func(), error) {
	if o.outPath == "" {
		return os.Stdout, func() {}, nil
	}
	f, err := os.Create(o.outPath)
	if err != nil {
		return nil, func() {}, fmt.Errorf("cannot open file %s: %w", o.outPath, err)
	}
	return f, func() { f.Close() }, nil
}
