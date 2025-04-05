package permissions

type PermissionDef struct {
	Name      string
	Operation PermissionType
}

func (p PermissionDef) Equals(p2 PermissionDef) bool {
	return p.Name == p2.Name && p.Operation == p2.Operation
}

func (p PermissionDef) ToString() string {
	return p.Operation.Val() + p.Name
}

type PermissionType string

const (
	PermissionRead   PermissionType = "read:"
	PermissionWrite  PermissionType = "write:"
	PermissionManage PermissionType = "manage:"
	PermissionDelete PermissionType = "delete:"
	PermissionCreate PermissionType = "create:"
	PermissionUpdate PermissionType = "update:"
	PermissionList   PermissionType = "list:"
)

func (p PermissionType) Val() string {
	return string(p)
}
