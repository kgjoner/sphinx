package identity

/* =============================================================================
	Welcome
============================================================================= */

type EmailWelcome struct {
	UserName string
	UserID   string
	Code     string
}

func (e EmailWelcome) TemplateKey() string {
	return "welcome"
}

/* =============================================================================
	Password Reset
============================================================================= */

type EmailResetPassword struct {
	UserName string
	UserID   string
	Code     string
}

func (e EmailResetPassword) TemplateKey() string {
	return "password_reset"
}

/* =============================================================================
	Password Update Notice
============================================================================= */

type EmailUpdatePasswordNotice struct {
	UserName string
}

func (e EmailUpdatePasswordNotice) TemplateKey() string {
	return "password_update_notice"
}

/* =============================================================================
	Email Update Notice
============================================================================= */

type EmailUpdateEmailNotice struct {
	UserName string
	NewEmail string
	UserID   string
}

func (e EmailUpdateEmailNotice) TemplateKey() string {
	return "email_update_notice"
}

/* =============================================================================
	Email Update Confirmation
============================================================================= */

type EmailConfirmEmailUpdate struct {
	UserName string
	NewEmail string
	UserID   string
	Code     string
}

func (e EmailConfirmEmailUpdate) TemplateKey() string {
	return "email_update_confirmation"
}
