terraform {
  required_providers {
    scylladb = {
      source = "registry.terraform.io/retailnext/scylladb"
    }
  }
}

provider "scylladb" {
  host = "localhost:9042"
  # port = 9042
  auth_login_userpass {
    username = "cassandra"
    password = "cassandra"
  }
  # tls_ca_cert = "ca.pem"
  # auth_tls {
  #   cert = 
  #   key = 
  # }
}

# resource "scyalldb_role" "admin" {
#   id = ""

# }

data "scylladb_role" "cassandra" {
  id = "cassandra"
}

output "cassandra_role" {
  value = data.scylladb_role.cassandra
}