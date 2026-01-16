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

func (c *Cluster) CreateRole(role Role) error {
	query := fmt.Sprintf("INSERT INTO %s.roles (role, can_login, is_superuser, member_of) VALUES (?, ?, ?, ?)", c.SystemAuthKeyspaceName)
	return c.Session.Query(query, role.Role, role.CanLogin, role.IsSuperuser, role.MemberOf).Exec()
}

func (c *Cluster) UpdateRole(role Role) error {
	query := fmt.Sprintf("UPDATE %s.roles SET can_login = ?, is_superuser = ?, member_of = ? WHERE role = ?", c.SystemAuthKeyspaceName)
	return c.Session.Query(query, role.CanLogin, role.IsSuperuser, role.MemberOf, role.Role).Exec()
}
