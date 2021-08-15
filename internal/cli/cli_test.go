package cli

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestRun(t *testing.T) {
	type output struct {
		path    string
		content string
	}
	type testCase struct {
		argv   []string
		out    *output
		stderr string
		status int
	}
	cases := []testCase{
		{[]string{"injecuet", "-output", "../../testdata/ok.out.cue", "../../testdata/ok.cue"}, &output{path: "../../testdata/ok.out.cue", content: "{\n\tname: string @injectenv(X_NAME)\n}\n"}, "", 0},
		{[]string{"injecuet"}, nil, "input file must be given\n", 1},
		{[]string{"injecuet", "../../testdata/ok.cue", "../../testdata/not_found.cue"}, nil, "input file must be given\n", 1},
		{[]string{"injecuet", "missing_file.cue"}, nil, "failed to inject values to file missing_file.cue: cannot parse file(missing_file.cue): open missing_file.cue: no such file or directory\n", 1},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("args=%s", strings.Join(tc.argv, " ")), func(t *testing.T) {
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
