package handlers

import (
	"fmt"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"github.com/labstack/echo/v4"
	"github.com/webpoint-solutions-llc/go-starter/internal/config"
	"github.com/webpoint-solutions-llc/go-starter/internal/constants"
	"github.com/webpoint-solutions-llc/go-starter/internal/dto"
	"github.com/webpoint-solutions-llc/go-starter/internal/errorhandler"
)

// Me godoc
// @Summary Get User Info
// @Description Get User Info
// @Tags User
// @Produce json
// @Success 200 {object} dto.UserInfoResponse "User Info"
// @Failure 400 {object} errorhandler.HttpError "Invalid request format"
// @Failure 500 {object} errorhandler.HttpError "Internal server error"
// @Security ApiKeyAuth
// @Security BearerAuth
// @Router /api/v1/user/me [get]
func (h *Handler) Me(c echo.Context) error {
	ctx := c.Request().Context()

	claims := c.Get("claims").(*dto.CustomClaims)

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return errorhandler.ErrorBadRequest(constants.MsgReLogin)
	}

	user, err := h.svc.GetUserByID(ctx, userID)
	if err != nil {
		return errorhandler.ErrorInternal(err)
	}

	userInfo := dto.UserInfoResponse{}
	copier.Copy(&userInfo, &user)
	return h.res.JSON(c, userInfo)
}

// ProfileImage godoc
// @Summary Upload profile image
// @Description Upload a profile image (JPEG/PNG/WebP up to 5MB) to object storage.
// @Tags User
// @Accept multipart/form-data
// @Produce json
// @Param data formData dto.ProfileImageUploadForm true "Multipart form data"
// @Success 200 {object} dto.ProfileImageUploadResponse "Successfully uploaded profile image"
// @Failure 400 {object} errorhandler.HttpError "Invalid request format or file"
// @Failure 500 {object} errorhandler.HttpError "Internal server error"
// @Security ApiKeyAuth
// @Security BearerAuth
// @Router /api/v1/user/upload-profile [post]
func (h *Handler) UploadProfileImage(c echo.Context) error {
	ctx := c.Request().Context()

	fh, err := c.FormFile("image")
	if err != nil {
		return errorhandler.ErrorBadRequest(constants.MsgNoImageSelected)
	}
	src, err := fh.Open()
	if err != nil {
		return errorhandler.ErrorInternal(err.Error())
	}
	defer src.Close()

	claims := c.Get("claims").(*dto.CustomClaims)

	ext := filepath.Ext(fh.Filename)
	// todo: use profile_pic in later implementation
	key := fmt.Sprintf("%s/%s/%s%s", config.ProfileDir, claims.Subject, uuid.New().String(), ext)

	if config.Cfg.IsMinio {
		key = fmt.Sprintf("/%s/%s/%s%s", config.ProfileDir, claims.Subject, uuid.New().String(), ext)
	}

	params := dto.S3UploadParams{
		Key:    key,
		Reader: src,
		Size:   fh.Size,
	}

	s3Upload, err := h.svc.UploadFile(ctx, params)
	if err != nil {
		return errorhandler.ErrorBadRequest(err)
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return errorhandler.ErrorBadRequest(constants.MsgReLogin)
	}

	update, err := h.svc.UpdateProfileImage(ctx, userID, s3Upload.Key)
	if err != nil {
		return errorhandler.ErrorInternal(err)
	}

	return h.res.JSON(c, update)
}
