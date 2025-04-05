package permissions

var (
	SecretCertAdmin = PermissionDef{"cert.admin", PermissionManage}
	SecretCertList  = PermissionDef{"cert.list", PermissionRead}

	SecretJwtAdmin  = PermissionDef{"jwt.admin", PermissionManage}
	SecretJwtNew    = PermissionDef{"jwt.new", PermissionWrite}
	SecretJwtVerify = PermissionDef{"jwt.verify", PermissionRead}
)

var (
	allPermissions = map[string]PermissionDef{}
)

func addPermissionDefMap(p PermissionDef) {
	allPermissions[p.ToString()] = p
}

func GetPermissionDef(s string) (PermissionDef, bool) {
	val, ok := allPermissions[s]
	return val, ok
}

func init() {
	allPermissions = map[string]PermissionDef{}
	addPermissionDefMap(SecretCertAdmin)
	addPermissionDefMap(SecretCertList)
	addPermissionDefMap(SecretJwtAdmin)
	addPermissionDefMap(SecretJwtNew)
	addPermissionDefMap(SecretJwtVerify)
}
