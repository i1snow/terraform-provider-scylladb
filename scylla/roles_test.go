package scylla

import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestGetRoleCassandra(t *testing.T) {
	devClusterHost := NewTestCluster(t)

	cluster := NewClusterConfig([]string{devClusterHost})
	cluster.SetSystemAuthKeyspace("system")
	if err := cluster.CreateSession(); err != nil {
		t.Fatalf("failed to create session: %s", err)
	}
	defer cluster.Session.Close()

	role, err := cluster.GetRole("cassandra")
	if err != nil {
		t.Fatalf("failed to get role: %s", err)
	}

	expectedRole := Role{
		Role:        "cassandra",
		CanLogin:    true,
		IsSuperuser: true,
		MemberOf:    []string{},
	}

	if role.Role != expectedRole.Role ||
		role.CanLogin != expectedRole.CanLogin ||
		role.IsSuperuser != expectedRole.IsSuperuser ||
		!equalStringSlices(role.MemberOf, expectedRole.MemberOf) {
		t.Fatalf("role does not match expected values. got: %+v, want: %+v", role, expectedRole)
	}
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

func NewTestCluster(t *testing.T) string {
	ctx := context.Background()
	scyllaDevContainer, err := testcontainers.Run(
		ctx, "scylladb/scylla:2025.4.1",
		testcontainers.WithCmdArgs("--developer-mode", "1", "--smp", "1", "--overprovisioned", "1"),
		testcontainers.WithExposedPorts("9042/tcp"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("9042/tcp"),
			// wait.ForLog("Ready to accept connections"),
		),
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
