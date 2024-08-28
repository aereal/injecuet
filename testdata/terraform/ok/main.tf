terraform {
  required_version = "~> 1.9.0"
}

output "user" {
  value = {
    name = "aereal"
    age  = 17
  }
}
