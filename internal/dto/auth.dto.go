package dto

type CodeExchangeRequest struct {
	Code string `json:"code" validate:"required"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IsFirstLogin bool   `json:"is_first_login,omitempty"`
	IsAdminMode  bool   `json:"is_admin,omitempty"`
}

type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type EmailLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	IsAdmin  bool   `json:"is_admin"`
}

type EmailSignUpRequest struct {
	Name                 string `json:"name" validate:"required,min=2,max=100"`
	PhoneNumber          string `json:"phone_number,omitempty" validate:"omitempty,e164"`
	Email                string `json:"email" validate:"required,email"`
	Password             string `json:"password" validate:"required,min=8"`
	PrivacyPolicy        bool   `json:"privacy_policy" validate:"required"`
	IsRegulationAccepted bool   `json:"is_regulation_accepted" validate:"required"`
}
type GoogleLoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	UserInfo     string `json:"user_info"`
}

type PasswordResetRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type PasswordResetTokenCheck struct {
	Token string `json:"token" validate:"required"`
}
type PasswordResetConfirmRequest struct {
	Password string `json:"password" validate:"required,min=8"`
	Token    string `json:"token" swaggerignore:"true"`
}

type ResendEmailVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type PasswordUpdateRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password" validate:"required"`
}
