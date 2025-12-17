package handlers

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/webpoint-solutions-llc/dba/internal/config"
	"github.com/webpoint-solutions-llc/dba/internal/constants"
	"github.com/webpoint-solutions-llc/dba/internal/db/sqlc"
	"github.com/webpoint-solutions-llc/dba/internal/dto"
	"github.com/webpoint-solutions-llc/dba/internal/errorhandler"
	"github.com/webpoint-solutions-llc/dba/internal/utils"
	"golang.org/x/sync/errgroup"
)

// MediaUpload godoc
// @Summary Upload media
// @Description Uploads a media file
// @Tags Media
// @Accept multipart/form-data
// @Produce json
// @Param data formData dto.MediaUploadForm true "Multipart form data"
// @Success 200 {object} dto.MediaUploadResponse "Successfully uploaded media file"
// @Failure 400 {object} errorhandler.HttpError "Invalid request format"
// @Failure 500 {object} errorhandler.HttpError "Internal server error"
// @Security ApiKeyAuth
// @Security BearerAuth
// @Router /api/v1/media/upload [post]
func (h *Handler) MediaUpload(c echo.Context) error {
	ctx := c.Request().Context()
	claims := c.Get("claims").(*dto.CustomClaims)

	var (
		mainUpload   *dto.S3FileUpload
		mainFilename string
		mainType     string
		mainFileId   string
	)

	form, err := c.MultipartForm()
	if err != nil {
		return errorhandler.ErrorBadRequest("invalid multipart form")
	}

	mainFile, mainExists := form.File["file"]
	thumbFile, thumbExists := form.File["thumbnail"]
	isPublic := false

	if val, ok := form.Value["isPublic"]; ok && len(val) > 0 && strings.ToLower(val[0]) == "true" {
		isPublic = true
	}

	if !mainExists || len(mainFile) == 0 {
		return errorhandler.ErrorBadRequest(constants.MsgNoFileSelected)
	}

	mainFH := mainFile[0] // First file (required)
	thumbFH := (*multipart.FileHeader)(nil)
	if thumbExists && len(thumbFile) > 0 {
		thumbFH = thumbFile[0] // Optional
	}

	g, gctx := errgroup.WithContext(ctx)

	mainType = utils.DetectMediaCategoryByExtension(mainFH.Filename)
	if mainType == "" {
		return errorhandler.ErrorBadRequest(constants.MsgFileFormatNotSupported)
	}
	mainFileId = uuid.New().String()
	baseDir := config.MediaDir
	if isPublic {
		baseDir = config.PublicDir // must be defined in your config
	}
	// Upload main file
	g.Go(func() error {
		src, err := mainFH.Open()
		if err != nil {
			return err
		}
		defer src.Close()

		ext := filepath.Ext(mainFH.Filename)
		key := fmt.Sprintf("%s/%s/%s%s", baseDir, claims.Subject, mainFileId, ext)
		if config.Cfg.IsMinio {
			key = fmt.Sprintf("/%s/%s/%s%s", baseDir, claims.Subject, mainFileId, ext)
		}

		params := dto.S3UploadParams{
			Key:         key,
			Reader:      src,
			Size:        mainFH.Size,
			ContentType: mainFH.Header.Get("Content-Type"),
		}

		upload, err := h.svc.UploadFile(gctx, params)
		if err != nil {
			return err
		}
		mainUpload = &upload
		mainFilename = mainFH.Filename
		return nil
	})

	// Upload thumbnail file (optional)
	g.Go(func() error {
		src, err := thumbFH.Open()
		if err != nil {
			return err
		}
		defer src.Close()

		ext := filepath.Ext(thumbFH.Filename)
		key := fmt.Sprintf("%s/%s/thumb_%s%s", config.MediaDir, claims.Subject, mainFileId, ext)
		if config.Cfg.IsMinio {
			key = fmt.Sprintf("/%s/%s/thumb_%s%s", config.MediaDir, claims.Subject, mainFileId, ext)
		}

		params := dto.S3UploadParams{
			Key:         key,
			Reader:      src,
			Size:        thumbFH.Size,
			ContentType: thumbFH.Header.Get("Content-Type"),
		}

		_, err = h.svc.UploadFile(gctx, params)
		if err != nil {
			return err
		}
		return nil
	})

	// Wait for both to finish
	if err := g.Wait(); err != nil {
		return errorhandler.ErrorInternal(err.Error())
	}

	// Save media metadata (only after both uploads succeed)
	mediaParams := sqlc.CreateMediaParams{
		Name: &mainFilename,
		Url:  &mainUpload.Key,
		Type: &mainType,
	}
	media, err := h.svc.CreateMediaMeta(ctx, mediaParams)
	if err != nil {
		return errorhandler.ErrorInternal(err)
	}

	resp := dto.MediaUploadResponse{
		Id:  media.ID.String(),
		Url: *mediaParams.Url,
	}

	return h.res.JSON(c, resp)
}

// MediaCleanupDelete godoc
// @Summary Delete media
// @Description Deletes a Unused media file
// @Tags Media
// @Accept json
// @Produce json
// @Failure 400 {object} errorhandler.HttpError "Invalid request or media ID"
// @Failure 404 {object} errorhandler.HttpError "Media not found"
// @Failure 500 {object} errorhandler.HttpError "Internal server error"
// @Security ApiKeyAuth
// @Router /api/v1/public/media-cleanup [delete]
func (h *Handler) MediaCleanup(c echo.Context) error {
	ctx := c.Request().Context()

	res, err := h.svc.GetUnusedMedia(ctx)
	if err != nil {
		return err
	}
	var (
		ids  []uuid.UUID
		urls []string
	)
	for _, media := range res {
		ids = append(ids, media.ID)

		if media.Url != nil {
			urls = append(urls, *media.Url)
		}
	}
	var objectsToDelete []s3types.ObjectIdentifier
	for _, url := range urls {
		objectsToDelete = append(objectsToDelete, s3types.ObjectIdentifier{
			Key: aws.String(url),
		})
	}
	bypassGovernance := false
	err = h.svc.DeleteObjects(ctx, objectsToDelete, bypassGovernance)
	if err != nil {
		return err
	}
	err = h.svc.DeleteMediaMetaByIdBulk(ctx, ids)
	if err != nil {
		return errorhandler.ErrorInternal(fmt.Errorf("failed to delete media metadata: %w", err))
	}

	return h.res.JSON(c, "Media Cleaned Successfully!")
}
