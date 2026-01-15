terraform {
  required_providers {
    scylladb = {
      source = "registry.terraform.io/retailnext/scylladb"
    }
  }
}

provider "scylladb" {
  host = "localhost:9042"
  port = 9042
  # auth_login_userpass = {
  #   username = "admin"
  #   password = "admin_password"
  # }
}

data "scylladb_role" "cassandra" {}

# output "cassandra_role" {
#   value = data.scylladb_role.cassandra
# }