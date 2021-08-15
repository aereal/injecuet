package injecuet

import (
	"fmt"
	"os"
	"testing"

	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/format"
	"github.com/google/go-cmp/cmp"
)

func TestInjectOK(t *testing.T) {
	type testCase struct {
		dataPath string
		want     string
		envs     map[string]string
	}
	cases := []testCase{
		{"./testdata/ok.cue", "name: \"aereal\" @injectenv(X_NAME)\n", map[string]string{"X_NAME": "aereal"}},
		{"./testdata/not_found.cue", "name: string @injectenv(X_UNKNOWN)\n", map[string]string{"X_NAME": "aereal"}},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("dataPath=%s", tc.dataPath), func(t *testing.T) {
			for name, value := range tc.envs {
				os.Setenv(name, value)
			}
			defer func() {
				for name := range tc.envs {
					os.Unsetenv(name)
				}
			}()

			injector := NewEnvironmentInjector()
			got, err := injector.Inject(tc.dataPath)
			if err != nil {
				t.Fatal(err)
			}
			cc := cuecontext.New()
			formattedWant, err := format.Node(cc.CompileString(tc.want).Syntax())
			if err != nil {
				t.Fatalf("cannot format want: %s", err)
			}
			formattedGot, err := format.Node(got.Syntax())
			if err != nil {
				t.Fatalf("cannot format got: %s", err)
			}
			if !cmp.Equal(string(formattedWant), string(formattedGot)) {
				diff := cmp.Diff(string(formattedWant), string(formattedGot))
				t.Errorf("(-want, +got):\n%s", diff)
			}
		})
	}
}
