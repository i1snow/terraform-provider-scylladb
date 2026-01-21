terraform {
  required_providers {
    scylladb = {
      source = "registry.terraform.io/retailnext/scylladb"
    }
  }
}

provider "scylladb" {
  host = "localhost:9042"
  auth_login_userpass {
    username = "cassandra"
    password = "cassandra"
  }
}

import {
  id = "admin"
  to = scylladb_role.admin
}

resource "scylladb_role" "admin" {
    role = "admin"
    can_login = false
    is_superuser = false
}

output "admin_role" {
  value = scylladb_role.admin
}