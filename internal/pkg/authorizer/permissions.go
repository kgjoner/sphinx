package authorizer

import (
	"github.com/kgjoner/sphinx/internal/domains/access"
	"github.com/kgjoner/sphinx/internal/domains/identity"
)

var appPermissions = []string{
	identity.PermUserReadAll,
	access.PermAppReadAll,
	access.PermLinkReadAll,
	access.PermLinkManage,
	access.PermRoleManageExtra,
}

var permissionsByRole = map[string][]string{
	"admin": {
		identity.PermUserWriteAll,
		identity.PermUserReadAll,
		access.PermAppCreate,
		access.PermAppReadAll,
		access.PermAppEdit,
		access.PermAppRecreateSecret,
		access.PermAppReadAll,
		access.PermLinkManage,
		access.PermLinkReadAll,
		access.PermRoleManageAdmin,
	},
	"manager": {
		identity.PermUserWriteAll,
		identity.PermUserReadAll,
		access.PermAppCreate,
		access.PermAppReadAll,
		access.PermAppEdit,
		access.PermAppReadAll,
		access.PermLinkManage,
		access.PermLinkReadAll,
		access.PermRoleManageManager,
	},
}
