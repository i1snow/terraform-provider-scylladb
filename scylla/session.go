// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package scylla

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type Cluster struct {
	Cluster                *gocql.ClusterConfig
	SystemAuthKeyspaceName string
	Session                *gocql.Session
}

func NewClusterConfig(hosts []string) Cluster {
	cluster := gocql.NewCluster(hosts...)
	cluster.DisableInitialHostLookup = true
	return Cluster{
		Cluster:                cluster,
		SystemAuthKeyspaceName: "system_auth",
	}
}

func (c *Cluster) CreateSession() error {
	session, err := c.Cluster.CreateSession()
	if err != nil {
		return err
	}
	c.Session = session
	return nil
}

func (c *Cluster) SetUserPasswordAuth(username, password string) {
	c.Cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: username,
		Password: password,
	}
}

func (c *Cluster) SetSystemAuthKeyspace(name string) {
	c.SystemAuthKeyspaceName = name
}

func GetTestCluster(t *testing.T) *Cluster {
	devClusterHost := NewTestCluster(t)

	cluster := NewClusterConfig([]string{devClusterHost})
	cluster.SetSystemAuthKeyspace("system")
	cluster.SetUserPasswordAuth("cassandra", "cassandra")
	if err := cluster.CreateSession(); err != nil {
		t.Fatalf("failed to create session: %s", err)
	}
	return &cluster
}

func NewTestCluster(t *testing.T) string {
	ctx := context.Background()

	// Get the config
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	scyllaConfig, err := filepath.Abs(filepath.Join(dir, "testdata", "scylla.yaml"))
	require.NoError(t, err)
	scyllaDevContainer, err := testcontainers.Run(
		ctx, "scylladb/scylla:2025.4.1",
		//testcontainers.WithCmdArgs("--developer-mode", "1", "--smp", "1", "--overprovisioned", "1"),
		testcontainers.WithCmdArgs("--smp", "1", "--overprovisioned", "1"),
		testcontainers.WithExposedPorts("9042/tcp"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("9042/tcp"),
			// wait.ForLog("Ready to accept connections"),
		),
		testcontainers.WithFiles(testcontainers.ContainerFile{
			HostFilePath:      scyllaConfig,
			ContainerFilePath: "/etc/scylla/scylla.yaml",
			FileMode:          0o777,
		}),
		// testcontainers.WithLogConsumerConfig(&testcontainers.LogConsumerConfig{
		// 	Opts:      []testcontainers.LogProductionOption{testcontainers.WithLogProductionTimeout(10 * time.Second)},
		// 	Consumers: []testcontainers.LogConsumer{&StdoutLogConsumer{}},
		// }),
	)
	if err != nil {
		t.Fatalf("failed to start the scylla container: %s", err)
	}

	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(scyllaDevContainer); err != nil {
			t.Fatalf("failed to terminate the scylla container: %s", err)
		}
	})

	host, err := scyllaDevContainer.PortEndpoint(ctx, "9042", "")
	if err != nil {
		t.Fatalf("failed to get the scylla container endpoint: %s", err)
	}
	return host
}

// StdoutLogConsumer is a LogConsumer that prints the log to stdout.
type StdoutLogConsumer struct{}

// Accept prints the log to stdout.
func (lc *StdoutLogConsumer) Accept(l testcontainers.Log) {
	fmt.Print(string(l.Content))
}
