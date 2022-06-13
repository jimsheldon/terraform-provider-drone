terraform {
  required_providers {
    drone = {
      version = "0.1"
      source = "jimsheldon.com/test/drone"
    }
  }
}

provider "drone" {}

data "drone_template" "base2" {
  name = "base2.yaml"
  namespace = "jimsheldon"
}

output "thing" {
  value = data.drone_template.base2.name
}