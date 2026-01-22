// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package scylla

import (
	"fmt"
	"strings"
)

type Keyspace struct {
	Name              string
	DurableWrites     bool
	ReplicationClass  string
	ReplicationFactor int
}

func (c *Cluster) GetKeyspace(keyspaceName string) (Keyspace, error) {
	var keyspace Keyspace
	var replication map[string]string

	query := "SELECT keyspace_name, durable_writes, replication FROM system_schema.keyspaces WHERE keyspace_name = ?"
	if err := c.Session.Query(query, keyspaceName).Scan(
		&keyspace.Name,
		&keyspace.DurableWrites,
		&replication,
	); err != nil {
		return Keyspace{}, err
	}

	keyspace.ReplicationClass = replication["class"]
	if rf, ok := replication["replication_factor"]; ok {
		_, _ = fmt.Sscanf(rf, "%d", &keyspace.ReplicationFactor)
	}

	return keyspace, nil
}

func (c *Cluster) CreateKeyspace(keyspace Keyspace) error {
	query := fmt.Sprintf(
		`CREATE KEYSPACE %s WITH replication = {'class': '%s', 'replication_factor': %d} AND durable_writes = %v`,
		quoteIdentifier(keyspace.Name),
		keyspace.ReplicationClass,
		keyspace.ReplicationFactor,
		keyspace.DurableWrites,
	)
	return c.Session.Query(query).Exec()
}

func (c *Cluster) UpdateKeyspace(keyspace Keyspace) error {
	query := fmt.Sprintf(
		`ALTER KEYSPACE %s WITH replication = {'class': '%s', 'replication_factor': %d} AND durable_writes = %v`,
		quoteIdentifier(keyspace.Name),
		keyspace.ReplicationClass,
		keyspace.ReplicationFactor,
		keyspace.DurableWrites,
	)
	return c.Session.Query(query).Exec()
}

func (c *Cluster) DeleteKeyspace(keyspace Keyspace) error {
	query := fmt.Sprintf(`DROP KEYSPACE %s`, quoteIdentifier(keyspace.Name))
	return c.Session.Query(query).Exec()
}

// quoteIdentifier quotes a CQL identifier to prevent injection.
// It doubles any existing double quotes and wraps the identifier in double quotes.
func quoteIdentifier(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}
