package main_test

import (
	"os"
	"os/exec"
)

func Example_tfstate() {
	cmd := exec.Command("go", "run", "github.com/aereal/injecuet/cmd/injecuet", "../../testdata/ok_tfstate.cue")
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		panic(err)
	}
	// Output: {
	// 	@inject(tfstate,stateURL="./terraform/ok/terraform.tfstate")
	// 	name: "aereal" @inject(tfstate,name="output.user.name")
	// 	age:  17       @inject(tfstate,name="output.user.age")
	// }
}
