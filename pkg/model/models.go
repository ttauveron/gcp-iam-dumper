package model

type Hierarchy struct {
	ID       string
	Name     string
	Type     string
	ParentID string
}

type Principal struct {
	ID   string
	Name string
	Type string
}

type PrincipalRelationship struct {
	ParentID string
	ChildID  string
}

type ResourceIAMPermission struct {
	ResourceID  string
	PrincipalID string
	RoleID      string
	Conditional string
	AssetType   string
	HierarchyID string
}

type Role struct {
	ID          string
	Title       string
	Permissions []string
}
