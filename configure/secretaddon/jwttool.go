package secretaddon

import (
	"github.com/meidoworks/nekoq-component/configure/secretapi"
)

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

type PermissionsList map[string]struct{}

type PermissionResourceList map[string]PermissionType

func (p PermissionResourceList) Dedup() PermissionsList {
	dedup := make(PermissionsList)
	for key, val := range p {
		dedup[string(val)+key] = struct{}{}
	}
	return dedup
}

const (
	JwtClaimsKeyPermissions = "permissions"
)

// PermissionOperator is an operator that performs permission checks on:
//   - matched permissions
//   - unmatched permissions: rest unmatched permissions in the target list
//   - extra permissions in the request list
type PermissionOperator func(matched, nonMatched, additional PermissionsList) bool

func AnyPermissionOperator(matched, nonMatched, additional PermissionsList) bool {
	return len(matched) > 0
}

func AllPermissionOperator(matched, nonMatched, additional PermissionsList) bool {
	return len(matched) > 0 && len(nonMatched) <= 0
}

type JwtTool struct {
}

func NewJwtTool() *JwtTool {
	return &JwtTool{}
}

func (j *JwtTool) SetupPermissions(data secretapi.JwtData, perms PermissionResourceList) {
	var list []string
	for key := range perms.Dedup() {
		list = append(list, key)
	}
	data[JwtClaimsKeyPermissions] = list
}

func (j *JwtTool) VerifyPermissions(data secretapi.JwtData, allowed PermissionResourceList, op PermissionOperator) bool {
	dedup := allowed.Dedup()

	// dedup request
	permissionField := data[JwtClaimsKeyPermissions].([]any)
	permissions := make(PermissionsList)
	for _, permission := range permissionField {
		permissions[permission.(string)] = struct{}{}
	}

	// matching
	matched := make(PermissionsList)
	for permission := range dedup {
		if _, ok := permissions[permission]; ok {
			delete(permissions, permission)
			delete(dedup, permission)
			matched[permission] = struct{}{}
		}
	}

	// operator apply
	return op(matched, dedup, permissions)
}
