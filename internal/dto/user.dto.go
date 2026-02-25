package dto

import (
	"io"
	"mime/multipart"
	"net/netip"

	"github.com/golang-jwt/jwt/v5"
)

type UserInfoResponse struct {
	ID                      string `json:"id"`
	Name                    string `json:"name"`
	Email                   string `json:"email"`
	PhoneNumber             string `json:"phone_number,omitempty"`
	Username                string `json:"username,omitempty"`
	ProfileImage            string `json:"profile_image,omitempty"`
	IsEmailVerified         bool   `json:"is_email_verified"`
	IsRegulationAccepted    bool   `json:"is_regulation_accepted"`
	IsPrivacyPolicyAccepted bool   `json:"is_privacy_policy_accepted"`
	AppleID                 string `json:"apple_id,omitempty"`
	GoogleID                string `json:"google_id,omitempty"`
	Role                    string `json:"role,omitempty"`
	Status                  string `json:"status,omitempty"`
	Profession              string `json:"profession,omitempty"`
	HasPassword             bool   `json:"has_password"`
	IsProfileComplete       bool   `json:"is_profile_complete"`
}
type ProfileImageUploadForm struct {
	Image *multipart.FileHeader `form:"image" validate:"required" swaggertype:"string" format:"binary"`
}

type ProfileImageUploadResponse struct {
	ProfileImage string `json:"profile_image"`
}

type EmailData struct {
	Name             string
	VerificationLink string
	SupportEmail     string
}

type PasswordResetData struct {
	Name              string
	PasswordResetLink string
	SupportEmail      string
}

type EmailVerificationData struct {
	Name             string
	VerificationLink string
	SupportEmail     string
}

type EmailDeactivate struct {
	Name         string
	LoginLink    string
	SupportEmail string
}
type SecondaryEmailVerificationData struct {
	Name         string
	OTP          string
	SupportEmail string
}

type LoginParams struct {
	// Common fields
	Email     string
	UserAgent *string
	IPAddress *netip.Addr
	IsAdmin   bool

	// Email login
	Password *string

	// Socail Login Common
	Name *string

	// Google login
	GoogleID *string

	// Apple login
	AppleID *string
}

type CheckoutSessionParams struct {
	CustomerID  string
	PriceID     string
	Quantity    int64
	Metadata    map[string]string
	PortfolioID string
}

type GoogleCallbackUserInfo struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	FamilyName    string `json:"family_name"`
	GivenName     string `json:"given_name"`
	Name          string `json:"name"`
	jwt.RegisteredClaims
}

type S3UploadParams struct {
	Key         string
	Reader      io.Reader
	Size        int64
	ContentType string
}

type S3FileUpload struct {
	Key         string
	Size        int64
	ContentType string
}

type ObjectIdentifier struct {
	Key       string
	VersionId *string // Use *string for optional fields to match AWS SDK's ObjectIdentifier
}

type BaiscUserInfo struct {
	Name             *string
	Username         *string
	PhoneNumber      *string
	Badge            *string
	StripeCustomerID *string
}

func (r BaiscUserInfo) IsEmpty() bool {
	return r.Name == nil &&
		r.Username == nil &&
		r.PhoneNumber == nil &&
		r.StripeCustomerID == nil
}

type UpdateUserInfoParams struct {
	User BaiscUserInfo
}
