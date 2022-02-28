package cli

import "io"

func Example_tfstate() {
	app := &App{errOut: io.Discard}
	_ = app.Run([]string{"injecuet", "../../testdata/ok_tfstate.cue"})
	// Output: {
	// 	@inject(tfstate,stateURL=./terraform/ok/terraform.tfstate)
	// 	name: "aereal" @inject(tfstate,name=output.user.name)
	// 	age:  17       @inject(tfstate,name=output.user.age)
	// }
}
