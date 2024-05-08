package db

import (
	"database/sql"
	"fmt"

	"github.com/ttauveron/gcp-iam-dumper/pkg/model"
)

func InsertHierarchies(db *sql.DB, hierarchies []model.Hierarchy) error {
	for _, hierarchy := range hierarchies {
		_, err := db.Exec(`INSERT INTO hierarchy (id, name, type, parent_id) VALUES (?, ?, ?, ?)`, hierarchy.ID, hierarchy.Name, hierarchy.Type, hierarchy.ParentID)
		if err != nil {
			return fmt.Errorf("error inserting hierarchy %v: %v", hierarchy, err)
		}
	}
	return nil
}
func InsertPrincipals(db *sql.DB, principals []model.Principal) error {
	stmt, err := db.Prepare("INSERT OR IGNORE INTO principal (id, name, type) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, p := range principals {
		_, err := stmt.Exec(p.ID, p.Name, p.Type)
		if err != nil {
			return err
		}
	}
	return nil
}

func InsertPrincipalRelationships(db *sql.DB, relationships []model.PrincipalRelationship) error {
	stmt, err := db.Prepare("INSERT INTO principal_hierarchy (parent_id, child_id) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, r := range relationships {
		_, err := stmt.Exec(r.ParentID, r.ChildID)
		if err != nil {
			return err
		}
	}
	return nil
}

func InsertResourceIAMPermission(db *sql.DB, permissions []model.ResourceIAMPermission) error {
	stmt, err := db.Prepare("INSERT INTO resource_role_principal (resource_id, principal_name, role_id, conditional, asset_type, hierarchy_id) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, permission := range permissions {
		_, err := stmt.Exec(permission.ResourceID, permission.PrincipalID, permission.RoleID, permission.Conditional, permission.AssetType, permission.HierarchyID)
		if err != nil {
			return err
		}
	}
	return nil
}

func InsertRoles(db *sql.DB, roles []model.Role) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	roleStmt, err := tx.Prepare("INSERT OR IGNORE INTO role (id, title) VALUES (?, ?)")
	if err != nil {
		return err
	}
	permissionStmt, err := tx.Prepare("INSERT OR IGNORE INTO role_permission (role_id, permission_id) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer roleStmt.Close()

	for _, r := range roles {
		for _, permission := range r.Permissions {
			_, err := permissionStmt.Exec(r.ID, permission)
			if err != nil {
				return err
			}
		}
		_, err := roleStmt.Exec(r.ID, r.Title)
		if err != nil {
			return err
		}

	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
