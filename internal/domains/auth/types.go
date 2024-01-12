package auth

/* ==============================================================================
	AccountCodeKind
============================================================================== */

type AccountCodeKind string

type accountCodeKind struct {	
	EMAIL_VERIFICATION AccountCodeKind
	PHONE_VERIFICATION AccountCodeKind
	PASSWORD_RESET     AccountCodeKind
}

func (s AccountCodeKind) Enumerate() any {
	return accountCodeKind{
		"email_verification",
		"phone_verification",
		"password_reset",
	}
}

var AccountCodeKindValues = AccountCodeKind.Enumerate("").(accountCodeKind)

/* ==============================================================================
	Roles
============================================================================== */

type Role string

type role struct {	
	ADMIN Role
	STAFF Role
}

func (s Role) Enumerate() any {
	return role{
		"ADMIN",
		"STAFF",
	}
}

var RoleValues = Role.Enumerate("").(role)