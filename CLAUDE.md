# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Development Commands

```bash
# Build provider
go build -v ./...

# Install provider locally
go install -v ./...

# Run linter (golangci-lint required)
golangci-lint run

# Format code
gofmt -s -w -e .

# Generate documentation (requires Terraform installed)
make generate

# Run unit tests
go test -v -cover -timeout=120s -parallel=10 ./...

# Run single test
go test -v -run TestFunctionName ./internal/provider/

# Run acceptance tests (creates real resources)
TF_ACC=1 go test -v -cover -timeout 120m ./...

# Default target (runs fmt, lint, install, generate)
make
```

## Architecture

This is a Terraform provider for ScyllaDB built on the Terraform Plugin Framework.

### Package Structure

- `main.go` - Entry point, configures provider server at `registry.terraform.io/retailnext/scylladb`
- `internal/provider/` - Terraform provider implementation
  - `provider.go` - Main provider configuration (host, port, auth)
  - `scylla_data_source.go` - Data source for reading ScyllaDB roles
  - `example_*.go` - Scaffolding examples for resources, data sources, functions, actions, ephemeral resources
- `scylla/` - ScyllaDB client wrapper
  - `session.go` - Cluster connection management using gocql
  - `roles.go` - Role queries against system_auth keyspace
- `internal/consts/` - Shared constants

### Provider Configuration

The provider connects to ScyllaDB using:
- `host` - ScyllaDB hostname (or `SCYLLADB_HOST` env var)
- `port` - ScyllaDB port
- `auth_login_userpass` - Optional username/password authentication

### Client Pattern

Resources and data sources receive `*scylla.Cluster` via `Configure()` method. The cluster wraps `gocql.ClusterConfig` and maintains a session.

## Linting Rules

The `.golangci.yml` enforces:
- Use `terraform-plugin-framework` instead of `terraform-plugin-sdk/v2`
- Use `terraform-plugin-testing` for test helpers instead of SDK equivalents
