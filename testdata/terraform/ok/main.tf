terraform {
  required_version = "~> 1.7.0"
}

output "user" {
  value = {
    name = "aereal"
    age  = 17
  }
}
