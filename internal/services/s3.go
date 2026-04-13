package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	s3Config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/webpoint-solutions-llc/go-starter/internal/config"
	"github.com/webpoint-solutions-llc/go-starter/internal/dto"
	"github.com/webpoint-solutions-llc/go-starter/internal/utils"
)

// S3Client wraps the AWS S3 client and presigner, and implements the storage interface.
type S3Client struct {
	client    *s3.Client
	presigner *s3.PresignClient
}

func newS3Client(ctx context.Context) (*S3Client, error) {
	cfg, err := s3Config.LoadDefaultConfig(ctx,
		s3Config.WithRegion(config.Cfg.S3Region),
	)
	if err != nil {
		return nil, err
	}

	svcOpts := []func(*s3.Options){}
	if config.Cfg.S3Endpoint != "" {
		svcOpts = append(svcOpts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(config.Cfg.S3Endpoint)
			o.UsePathStyle = true
		})
	}

	client := s3.NewFromConfig(cfg, svcOpts...)
	return &S3Client{
		client:    client,
		presigner: s3.NewPresignClient(client),
	}, nil
}

func (c *S3Client) UploadFile(ctx context.Context, params dto.S3UploadParams) (dto.S3FileUpload, error) {
	uploadCtx, cancel := context.WithTimeout(ctx, 4*time.Minute)
	defer cancel()

	uploader := manager.NewUploader(c.client, func(u *manager.Uploader) {
		u.PartSize = 5 * 1024 * 1024
		u.Concurrency = 3
	})

	contentType := params.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err := uploader.Upload(uploadCtx, &s3.PutObjectInput{
		Bucket:      aws.String(config.Cfg.S3Bucket),
		Key:         aws.String(params.Key),
		Body:        params.Reader,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return dto.S3FileUpload{}, fmt.Errorf("failed to upload file to S3: %w", err)
	}

	return dto.S3FileUpload{
		Key:         params.Key,
		Size:        params.Size,
		ContentType: contentType,
	}, nil
}

func (c *S3Client) S3MediaURL(ctx context.Context, key string) (string, error) {
	if config.Cfg.S3Endpoint != "" {
		return utils.JoinS3URL(key)
	}

	expires, err := time.ParseDuration(config.Cfg.S3SignedURLDuration)
	if err != nil {
		expires = 3 * time.Hour
	}

	resp, err := c.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(config.Cfg.S3Bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expires))
	if err != nil {
		return "", err
	}
	return resp.URL, nil
}

func (c *S3Client) DeleteObjects(ctx context.Context, objects []s3types.ObjectIdentifier, bypassGovernance bool) error {
	deleteCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	if len(objects) == 0 {
		return nil
	}

	awsObjects := make([]s3types.ObjectIdentifier, len(objects))
	for i, obj := range objects {
		awsObjects[i] = s3types.ObjectIdentifier{
			Key:       aws.String(*obj.Key),
			VersionId: obj.VersionId,
		}
	}

	input := s3.DeleteObjectsInput{
		Bucket: aws.String(config.Cfg.S3Bucket),
		Delete: &s3types.Delete{
			Objects: awsObjects,
			Quiet:   aws.Bool(true),
		},
	}
	if bypassGovernance {
		input.BypassGovernanceRetention = aws.Bool(true)
	}

	delOut, err := c.client.DeleteObjects(deleteCtx, &input)
	if err != nil {
		var noBucket *s3types.NoSuchBucket
		if errors.As(err, &noBucket) {
			log.Printf("Error deleting objects: Bucket %s does not exist.\n", config.Cfg.S3Bucket)
			return fmt.Errorf("failed to delete objects: %w", noBucket)
		}
		return fmt.Errorf("failed to delete objects from bucket %s: %w", config.Cfg.S3Bucket, err)
	}

	for _, outErr := range delOut.Errors {
		log.Printf("Failed to delete object %s: %s\n", *outErr.Key, *outErr.Message)
	}

	for _, delObjs := range delOut.Deleted {
		log.Printf("Successfully requested deletion of %s.\n", *delObjs.Key)
		waitErr := s3.NewObjectNotExistsWaiter(c.client).Wait(
			deleteCtx, &s3.HeadObjectInput{Bucket: aws.String(config.Cfg.S3Bucket), Key: delObjs.Key}, time.Minute)
		if waitErr != nil {
			log.Printf("Warning: Failed to confirm deletion of object %s: %v\n", *delObjs.Key, waitErr)
		} else {
			log.Printf("Confirmed deletion of %s.\n", *delObjs.Key)
		}
	}
	return nil
}

// The following methods delegate to s.store so that Service satisfies the handler-level
// MediaService and UserService interfaces while keeping S3 logic behind the storage interface.

func (s *Service) UploadFile(ctx context.Context, params dto.S3UploadParams) (dto.S3FileUpload, error) {
	return s.store.UploadFile(ctx, params)
}

func (s *Service) S3MediaURL(ctx context.Context, key string) (string, error) {
	return s.store.S3MediaURL(ctx, key)
}

func (s *Service) DeleteObjects(ctx context.Context, objects []s3types.ObjectIdentifier, bypassGovernance bool) error {
	return s.store.DeleteObjects(ctx, objects, bypassGovernance)
}
