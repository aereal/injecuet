package injecuet

import (
	"fmt"
	"os"
	"strings"
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
		match    func(name string) bool
	}
	cases := []testCase{
		{"./testdata/ok.cue", "name: \"aereal\" @inject(env,name=X_NAME)\n", map[string]string{"X_NAME": "aereal"}, matchAll},
		{"./testdata/ok_deprecated.cue", "name: \"aereal\" @injectenv(X_NAME)\n", map[string]string{"X_NAME": "aereal"}, matchAll},
		{"./testdata/not_found.cue", "name: string @inject(env,name=X_UNKNOWN)\n", map[string]string{"X_NAME": "aereal"}, matchAll},
		{"./testdata/partial.cue", "{\n\tname: string @inject(env,name=X_NAME)\n\tage:  \"17\"  @inject(env,name=X_AGE)\n}", map[string]string{"X_NAME": "aereal", "X_AGE": "17"}, func(name string) bool { return strings.HasSuffix(name, "AGE") }},
		{"./testdata/hidden1.cue", "_name: \"aereal\" @inject(env,name=X_NAME)\n", map[string]string{"X_NAME": "aereal"}, matchAll},
		{"./testdata/hidden2.cue", "#name: \"aereal\" @inject(env,name=X_NAME)\n", map[string]string{"X_NAME": "aereal"}, matchAll},
		{"./testdata/missing_kind.cue", "name: string @inject(name=X_NAME)\n", map[string]string{"X_NAME": "aereal"}, matchAll},
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

			injector := NewInjector(NewEnvFillter(tc.match))
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

func TestInject_tfstate(t *testing.T) {
	type testCase struct {
		dataPath    string
		want        string
		tfstatePath string
	}
	cases := []testCase{
		{"./testdata/ok_tfstate.cue", "{\n\t@inject(tfstate,stateURL=./terraform/ok/terraform.tfstate)\n\tname: \"aereal\" @inject(tfstate,name=output.user.name)\n\tage:  17       @inject(tfstate,name=output.user.age)\n}", "./testdata/terraform/ok/terraform.tfstate"},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("dataPath=%s tfstatePath=%s", tc.dataPath, tc.tfstatePath), func(t *testing.T) {
			injector := NewInjector(NewTFStateFiller())
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
