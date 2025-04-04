package permissions

type PermissionDef struct {
	Name      string
	Operation PermissionType
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
