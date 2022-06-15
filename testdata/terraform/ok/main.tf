terraform {
  required_version = "~> 1.2.0"
}

output "user" {
  value = {
    name = "aereal"
    age  = 17
  }
}
