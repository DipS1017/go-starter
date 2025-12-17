package types

type TokenType string

const (
	TokenTypeAccess        TokenType = "access"
	TokenTypeRefresh       TokenType = "refresh"
	TokenTypeVerify        TokenType = "verify"
	TokenTypePasswordReset TokenType = "password_reset"
)

type Role string

const (
	RoleTypeAdmin     Role = "ADMIN"
	RoleTypeUser      Role = "USER"
	RoleTypeAdminMode Role = "ADMIN_MODE"
	RoleTypeModerator Role = "MODERATOR"
)
