package dto

import (
	"github.com/google/uuid"
	"github.com/webpoint-solutions-llc/go-starter/internal/db/sqlc"
)

type SecondaryEmailRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password"`
}

type ResendVerifyLinkEmailRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type EmailVerifyRequest struct {
	Otp string `json:"otp" validate:"required"`
}
type SwitchPrimaryEmailRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type SecondaryEmailResponse struct {
	ID         uuid.UUID `json:"id"`
	Email      string    `json:"email"`
	IsVerified bool      `json:"is_verified"`
}

type IDVerificationData struct {
	Name             string
	VerificationLink string
	JobLink          string
	JobTitle         string
	SupportEmail     string
	RejectReason     string
	JobDescription   string
}

type SummaryEmailTemplate struct {
	Recipient           sqlc.GetUserByIDRow
	SupportEmail        string
	SettingsURL         string
	FrontendURL         string
	SummaryNotification map[string]int
}
type EmailNotification struct {
	RecipientID      uuid.UUID
	SenderID         uuid.UUID
	NotificationType string
	TemplateName     string
	Subject          string
	Data             any
}
