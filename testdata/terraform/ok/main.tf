terraform {
  required_version = "~> 1.1.0"
}

output "user" {
  value = {
    name = "aereal"
    age  = 17
  }
}
