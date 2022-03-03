@inject(tfstate,stateURL="../testdata/terraform/ok/terraform.tfstate")

name: string @inject(tfstate,name="output.user.name")
age: int @inject(tfstate,name="output.user.age")
