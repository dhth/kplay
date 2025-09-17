package aws

import (
	"context"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type UploadResult struct {
	ObjectKey string
	Err       error
}

func UploadToS3(ctx context.Context, resultChan chan<- UploadResult, client *s3.Client, body io.Reader, bucketName, objectKey string) {
	result := UploadResult{
		ObjectKey: objectKey,
	}

	for range 5 {
		uploadCtx, uploadCancel := context.WithTimeout(ctx, 5*time.Second)
		result.Err = uploadToS3(uploadCtx, client, body, bucketName, objectKey)
		uploadCancel()

		select {
		case <-ctx.Done():
			return
		default:
			if result.Err == nil {
				resultChan <- result
				return
			}
		}
	}

	resultChan <- result
}

func uploadToS3(ctx context.Context, client *s3.Client, body io.Reader, bucketName, objectKey string) error {
	_, err := client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Body:   body,
	})

	return err
}
