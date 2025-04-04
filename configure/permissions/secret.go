package permissions

var (
	SecretCertAdmin = PermissionDef{"cert.admin", PermissionManage}
	SecretCertList  = PermissionDef{"cert.list", PermissionRead}
)
