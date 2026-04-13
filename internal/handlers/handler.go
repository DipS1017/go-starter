package handlers

import (
	"context"
	"log/slog"

	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
	"github.com/webpoint-solutions-llc/go-starter/internal/db/sqlc"
	"github.com/webpoint-solutions-llc/go-starter/internal/dto"
	"github.com/webpoint-solutions-llc/go-starter/internal/interfaces"
	"github.com/webpoint-solutions-llc/go-starter/internal/responders"
	"github.com/webpoint-solutions-llc/go-starter/internal/services"
	"github.com/webpoint-solutions-llc/go-starter/internal/types"
)

// AuthService covers all authentication and session operations used by auth handlers.
type AuthService interface {
	EmailLogin(ctx context.Context, params dto.LoginParams) (dto.LoginResponse, error)
	EmailSignUp(ctx context.Context, params dto.EmailSignUpRequest) (string, error)
	InserSocialLoginUser(ctx context.Context, params dto.LoginParams) (dto.LoginResponse, error)
	VerifyUserEmail(ctx context.Context, token string) (sqlc.UpdateEmailStatusRow, error)
	PasswordResetLink(ctx context.Context, email string) (sqlc.CreateResetTokenRow, error)
	PasswordResetTokenCheck(ctx context.Context, params dto.PasswordResetTokenCheck) (sqlc.UpdatePasswordResetRow, error)
	PasswordResetConfirm(ctx context.Context, params dto.PasswordResetConfirmRequest) (sqlc.UpdatePasswordResetRow, error)
	ResendEmailVerification(ctx context.Context, email string) (string, error)
	UpdatePassword(ctx context.Context, params dto.PasswordUpdateRequest, userID uuid.UUID) error
	GetSessionByRefreshTokenHash(ctx context.Context, hashedToken string) (sqlc.GetSessionByRefreshTokenHashRow, error)
	CleanupSession(ctx context.Context, sessionID uuid.UUID) (*uuid.UUID, error)
	GetAppleLoginURL() (string, error)
	ExchangeCodeWithApple(ctx context.Context, code string) (*types.AppleIDTokenClaims, error)
}

// UserService covers user profile and file-upload operations used by user handlers.
type UserService interface {
	GetUserByID(ctx context.Context, userID uuid.UUID) (sqlc.GetUserByIDRow, error)
	UpdateProfileImage(ctx context.Context, userID uuid.UUID, imagePath string) (sqlc.UpdateProfileImageRow, error)
	UploadFile(ctx context.Context, params dto.S3UploadParams) (dto.S3FileUpload, error)
}

// MediaService covers media upload, cleanup, and deletion operations.
type MediaService interface {
	UploadFile(ctx context.Context, params dto.S3UploadParams) (dto.S3FileUpload, error)
	CreateMediaMeta(ctx context.Context, params sqlc.CreateMediaParams) (sqlc.CreateMediaRow, error)
	GetUnusedMedia(ctx context.Context) ([]sqlc.GetUnusedMediaRow, error)
	DeleteObjects(ctx context.Context, objects []s3types.ObjectIdentifier, bypassGovernance bool) error
	DeleteMediaMetaByIDBulk(ctx context.Context, params []uuid.UUID) error
}

// WebhookService covers operations triggered by external webhook events.
type WebhookService interface {
	UpdateUser(ctx context.Context, userID uuid.UUID, userInfo dto.UpdateUserInfoParams) error
}

type Handler struct {
	auth    AuthService
	user    UserService
	media   MediaService
	webhook WebhookService
	log     *slog.Logger
	res     interfaces.Responders
}

func NewHandler() *Handler {
	svc := services.NewService()
	logger := slog.Default().With("component", "handler")
	return &Handler{
		auth:    svc,
		user:    svc,
		media:   svc,
		webhook: svc,
		log:     logger,
		res:     responders.NewResponder(logger),
	}
}
