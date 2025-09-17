package awsweb

import (
	"context"
	"io"
	"log/slog"
	"math/rand/v2"
	"time"

	// "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type UploadResult struct {
	ObjectKey string
	Err       error
}

func UploadToS3(_ context.Context, resultChan chan<- UploadResult, _ *s3.Client, _ io.Reader, _, objectKey string) {
	slog.Info("uploading to s3", "object_key", objectKey)
	mi := 300
	mx := 1500
	sleepMillis := rand.IntN(mx-mi+1) + mi
	time.Sleep(time.Duration(sleepMillis) * time.Millisecond)

	result := UploadResult{
		ObjectKey: objectKey,
	}
	resultChan <- result

	//
	// for i := range 5 {
	// 	uploadCtx, uploadCancel := context.WithTimeout(ctx, 5*time.Second)
	// 	result.Err = upload(uploadCtx, client, body, bucketName, objectKey)
	// 	uploadCancel()
	//
	// 	if result.Err != nil {
	// 		slog.Info("uploading to s3 failed", "object_key", objectKey, "attempt_num", i)
	// 	}
	//
	// 	select {
	// 	case <-ctx.Done():
	// 		return
	// 	default:
	// 		if result.Err == nil {
	// 			slog.Info("uploaded to s3", "object_key", objectKey, "attempt_num", i)
	// 			resultChan <- result
	// 			return
	// 		}
	// 	}
	// }
	//
	// resultChan <- result
}

//
// func upload(ctx context.Context, client *s3.Client, body io.Reader, bucketName, objectKey string) error {
// 	_, err := client.PutObject(ctx, &s3.PutObjectInput{
// 		Bucket: aws.String(bucketName),
// 		Key:    aws.String(objectKey),
// 		Body:   body,
// 	})
//
// 	return err
// }
