package scylla

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestGetRoleCassandra(t *testing.T) {
	cluster := GetTestCluster(t)
	defer cluster.Session.Close()

	role, err := cluster.GetRole("cassandra")
	if err != nil {
		t.Fatalf("failed to get role: %s", err)
	}

	expectedRole := Role{
		Role:        "cassandra",
		CanLogin:    true,
		IsSuperuser: true,
		MemberOf:    nil,
	}

	assert.Equal(t, expectedRole, role)
}

func TestCreateRole(t *testing.T) {
	cluster := GetTestCluster(t)
	defer cluster.Session.Close()

	inputRole := Role{
		Role: "testRole",
	}
	expectedRole := Role{
		Role:        "testRole",
		IsSuperuser: false,
		CanLogin:    false,
		MemberOf:    nil,
	}

	err := cluster.CreateRole(inputRole)
	if err != nil {
		t.Fatalf("failed to create a role: %s", err)
	}

	role, err := cluster.GetRole(inputRole.Role)
	if err != nil {
		t.Fatalf("failed to get a role for %s: %s", inputRole.Role, err)
	}

	assert.Equal(t, expectedRole, role)
}

func TestUpdateRole(t *testing.T) {
	cluster := GetTestCluster(t)
	defer cluster.Session.Close()

	inputRole := Role{
		Role: "testRole",
	}
	updateRole := Role{
		Role:        "testRole",
		IsSuperuser: true,
		CanLogin:    true,
	}
	expectedRole := Role{
		Role:        "testRole",
		IsSuperuser: true,
		CanLogin:    true,
		MemberOf:    nil,
	}

	err := cluster.CreateRole(inputRole)
	if err != nil {
		t.Fatalf("failed to create a role: %s", err)
	}

	err = cluster.UpdateRole(updateRole)
	if err != nil {
		t.Fatalf("failed to update a role: %s", err)
	}

	role, err := cluster.GetRole(inputRole.Role)
	if err != nil {
		t.Fatalf("failed to get a role for %s: %s", inputRole.Role, err)
	}

	assert.Equal(t, expectedRole, role)
}

func TestDeleteRole(t *testing.T) {
	cluster := GetTestCluster(t)
	defer cluster.Session.Close()

	inputRole := Role{
		Role: "testRole",
	}

	err := cluster.CreateRole(inputRole)
	if err != nil {
		t.Fatalf("failed to create a role: %s", err)
	}

	role, err := cluster.GetRole(inputRole.Role)
	if err != nil {
		t.Fatalf("failed to get a role for %s: %s", inputRole.Role, err)
	}
	fmt.Println(role)

	err = cluster.DeleteRole(inputRole)
	if err != nil {
		t.Fatalf("failed to delete a role: %s", err)
	}

	_, err = cluster.GetRole(inputRole.Role)
	assert.EqualError(t, err, "not found")
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	stringMap := make(map[string]bool)
	for _, str := range a {
		stringMap[str] = true
	}
	for _, str := range b {
		if !stringMap[str] {
			return false
		}
	}
	return true
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
	scyllaConfig, err := filepath.Abs(filepath.Join(".", "testdata", "scylla.yaml"))
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

// StdoutLogConsumer is a LogConsumer that prints the log to stdout
type StdoutLogConsumer struct{}

// Accept prints the log to stdout
func (lc *StdoutLogConsumer) Accept(l testcontainers.Log) {
	fmt.Print(string(l.Content))
}
