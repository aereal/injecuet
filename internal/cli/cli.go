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

	out, close, err := openOutput(outPath)
	if err != nil {
		fmt.Fprintln(a.errOut, err.Error())
		return 1
	}
	defer close()
	var match func(name string) bool
	if pattern != "" {
		re, err := regexp.Compile(pattern)
		if err != nil {
			fmt.Fprintf(a.errOut, "cannot parse pattern: %s\n", err)
			return 1
		}
		match = re.MatchString
	}
	if err := a.runMain(fs.Arg(0), out, match); err != nil {
		fmt.Fprintf(a.errOut, "%v\n", err)
		return 1
	}

	return 0
}

func (a *App) runMain(src string, out io.Writer, match func(name string) bool) error {
	injector := injecuet.NewInjector(injecuet.WithEnvironmentVariables(match), injecuet.WithTFState())
	v, err := injector.Inject(src)
	if err != nil {
		return fmt.Errorf("failed to inject values to file %s: %w", src, err)
	}
	formatted, err := format.Node(v.Syntax())
	if err != nil {
		return fmt.Errorf("failed to format file %s: %w", src, err)
	}
	_, _ = out.Write(formatted)
	_, _ = out.Write([]byte{'\n'})
	return nil
}

func openOutput(outPath string) (io.Writer, func(), error) {
	if outPath == "" {
		return os.Stdout, func() {}, nil
	}
	f, err := os.Create(outPath)
	if err != nil {
		return nil, func() {}, fmt.Errorf("cannot open file %s: %w", outPath, err)
	}
	return f, func() { f.Close() }, nil
}
