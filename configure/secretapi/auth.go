package secretapi

type AuthProvider interface {
	CheckOperationPermission() (bool, error)
}
