package scylla

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
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
