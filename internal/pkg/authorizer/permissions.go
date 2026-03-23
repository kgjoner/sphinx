package authorizer

import (
	"github.com/kgjoner/sphinx/internal/domains/access"
	"github.com/kgjoner/sphinx/internal/domains/identity"
)

var appPermissions = []string{
	identity.PermUserReadAll,
	identity.PermUserList,
	access.PermAppReadAll,
	access.PermAppEdit,
	access.PermAppRecreateSecret,
	access.PermLinkReadAll,
	access.PermLinkManage,
	access.PermRoleManageExtra,
	access.PermRoleManageManager,
}

var permissionsByRole = map[string][]string{
	"admin": {
		identity.PermUserWriteAll,
		identity.PermUserReadAll,
		identity.PermUserList,
		access.PermAppCreate,
		access.PermAppReadAll,
		access.PermAppEdit,
		access.PermAppRecreateSecret,
		access.PermAppReadAll,
		access.PermLinkManage,
		access.PermLinkReadAll,
		access.PermRoleManageExtra,
		access.PermRoleManageManager,
		access.PermRoleManageAdmin,
	},
	"manager": {
		identity.PermUserWriteAll,
		identity.PermUserReadAll,
		identity.PermUserList,
		access.PermAppCreate,
		access.PermAppReadAll,
		access.PermAppEdit,
		access.PermAppReadAll,
		access.PermLinkManage,
		access.PermLinkReadAll,
		access.PermRoleManageManager,
	},
}
