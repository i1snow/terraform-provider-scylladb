package scylla

import "fmt"

type Role struct {
	Role        string
	CanLogin    bool
	IsSuperuser bool
	MemberOf    []string
}

func (c *Cluster) GetRole(roleName string) (Role, error) {
	var role Role
	query := fmt.Sprintf("SELECT role, can_login, is_superuser, member_of FROM %s.roles WHERE role = ?", c.SystemAuthKeyspaceName)
	if err := c.Session.Query(query, roleName).Scan(
		&role.Role,
		&role.CanLogin,
		&role.IsSuperuser,
		&role.MemberOf,
	); err != nil {
		return Role{}, err
	}
	return role, nil
}
