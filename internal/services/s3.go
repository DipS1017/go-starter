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

func newS3Client(ctx context.Context) (*S3Client, error) {
	// only pass region into LoadDefaultConfig
	cfg, err := s3Config.LoadDefaultConfig(ctx,
		s3Config.WithRegion(config.Cfg.S3Region),
	)
	if err != nil {
		return nil, err
	}

	// build service‐specific options
	svcOpts := []func(*s3.Options){}

	if config.Cfg.S3Endpoint != "" {
		// in dev: point at MinIO and force path‐style URLs
		svcOpts = append(svcOpts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(config.Cfg.S3Endpoint)
			o.UsePathStyle = true
		})
	}

	// instantiate client with our options
	client := s3.NewFromConfig(cfg, svcOpts...)
	presigner := s3.NewPresignClient(client)

	return &S3Client{
		client:    client,
		presigner: presigner,
	}, nil
}

func (s *Service) UploadFile(ctx context.Context, params dto.S3UploadParams) (dto.S3FileUpload, error) {
	uploadCtx, cancel := context.WithTimeout(ctx, 4*time.Minute)
	defer cancel()

	uploader := manager.NewUploader(s.s3Client.client, func(u *manager.Uploader) {
		u.PartSize = 5 * 1024 * 1024 // 5 MiB per part
		u.Concurrency = 3
	})

	// Detect content type if not provided
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

	s3Upload := dto.S3FileUpload{
		Key:         params.Key,
		Size:        params.Size,
		ContentType: contentType,
	}

	return s3Upload, nil
}

func (s *Service) S3MediaURL(ctx context.Context, key string) (string, error) {
	// if its using minio
	if config.Cfg.S3Endpoint != "" {
		return utils.JoinS3URL(key)
	}

	expires, err := time.ParseDuration(config.Cfg.S3SignedURLDuration)
	if err != nil {
		expires = 3 * time.Hour
	}

	resp, err := s.s3Client.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(config.Cfg.S3Bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expires))
	if err != nil {
		return "", err
	}
	return resp.URL, nil
}

// DeleteObjects deletes a list of objects from the configured S3 bucket.
func (s *Service) DeleteObjects(ctx context.Context, objects []s3types.ObjectIdentifier, bypassGovernance bool) error {
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

	delOut, err := s.s3Client.client.DeleteObjects(deleteCtx, &input)
	if err != nil {
		var noBucket *s3types.NoSuchBucket
		if errors.As(err, &noBucket) {
			log.Printf("Error deleting objects: Bucket %s does not exist.\n", config.Cfg.S3Bucket)
			return fmt.Errorf("failed to delete objects: %w", noBucket)
		}
		return fmt.Errorf("failed to delete objects from bucket %s: %w", config.Cfg.S3Bucket, err)
	}

	if len(delOut.Errors) > 0 {
		for _, outErr := range delOut.Errors {
			log.Printf("Failed to delete object %s: %s\n", *outErr.Key, *outErr.Message)
		}
	}

	for _, delObjs := range delOut.Deleted {
		log.Printf("Successfully requested deletion of %s.\n", *delObjs.Key)
		waitErr := s3.NewObjectNotExistsWaiter(s.s3Client.client).Wait(
			deleteCtx, &s3.HeadObjectInput{Bucket: aws.String(config.Cfg.S3Bucket), Key: delObjs.Key}, time.Minute)
		if waitErr != nil {
			log.Printf("Warning: Failed to confirm deletion of object %s: %v\n", *delObjs.Key, waitErr)
		} else {
			log.Printf("Confirmed deletion of %s.\n", *delObjs.Key)
		}
	}
	return nil
}
