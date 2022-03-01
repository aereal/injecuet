package cli

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestRun(t *testing.T) {
	type output struct {
		path    string
		content string
	}
	type testCase struct {
		name   string
		argv   []string
		out    *output
		stderr string
		status int
	}
	cases := []testCase{
		{"ok", []string{"injecuet", "-output", "../../testdata/ok.out.cue", "../../testdata/ok.cue"}, &output{path: "../../testdata/ok.out.cue", content: "{\n\tname: string @inject(env,name=X_NAME)\n}\n"}, "", 0},
		{"no input file", []string{"injecuet"}, nil, "input file must be given\n", 1},
		{"refer missing environment variable", []string{"injecuet", "../../testdata/ok.cue", "../../testdata/not_found.cue"}, nil, "input file must be given\n", 1},
		{"input file not found", []string{"injecuet", "missing_file.cue"}, nil, "failed to inject values to file missing_file.cue: cannot parse file(missing_file.cue): open missing_file.cue: no such file or directory\n", 1},
		{"corrupt pattern", []string{"injecuet", "-pattern", "[", "-output", "../../testdata/ok.out.cue", "../../testdata/ok.cue"}, nil, "cannot parse pattern: error parsing regexp: missing closing ]: `[`\n", 1},
		{"ok; with pattern", []string{"injecuet", "-pattern", "AGE$", "-output", "../../testdata/partial.out.cue", "../../testdata/partial.cue"}, &output{path: "../../testdata/partial.out.cue", content: "{\n\tname: string @inject(env,name=X_NAME)\n\tage:  string @inject(env,name=X_AGE)\n}\n"}, "", 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			errOut := new(bytes.Buffer)
			app := &App{errOut: errOut}
			got := app.Run(tc.argv)
			if got != tc.status {
				t.Errorf("status: want=%d got=%d", tc.status, got)
			}
			gotErr := errOut.String()
			if gotErr != tc.stderr {
				t.Errorf("stderr (-want, +got):\n%s", cmp.Diff(tc.stderr, gotErr))
			}

			if tc.out == nil {
				return
			}
			defer func(outPath string) {
				if outPath == "" {
					return
				}
				os.Remove(outPath)
			}(tc.out.path)
			out, err := ioutil.ReadFile(tc.out.path)
			if err != nil {
				t.Fatalf("cannot read file (%s): %s", tc.out.path, err)
			}
			outStr := string(out)
			if outStr != tc.out.content {
				t.Errorf("output (-want, +got):\n%s", cmp.Diff(tc.out.content, outStr))
			}
		})
	}
}
