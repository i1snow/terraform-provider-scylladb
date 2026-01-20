package scylla

import (
	"fmt"
)

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

func (c *Cluster) CreateRole(role Role) error {
	query := fmt.Sprintf(`CREATE ROLE '%s' WITH LOGIN = %v AND SUPERUSER = %v`, role.Role, role.CanLogin, role.IsSuperuser)
	return c.Session.Query(query, role.CanLogin, role.IsSuperuser).Exec()
}

func (c *Cluster) UpdateRole(role Role) error {
	query := fmt.Sprintf(`ALTER ROLE '%s' WITH LOGIN = %v AND SUPERUSER = %v`, role.Role, role.CanLogin, role.IsSuperuser)
	return c.Session.Query(query, role.CanLogin, role.IsSuperuser).Exec()
}

func (c *Cluster) DeleteRole(role Role) error {
	query := fmt.Sprintf(`DROP ROLE '%s'`, role.Role)
	return c.Session.Query(query).Exec()
}
