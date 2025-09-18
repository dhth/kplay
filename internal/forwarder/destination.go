package forwarder

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Destination interface {
	Display() string
	upload(ctx context.Context, body io.Reader, fileName string) error
	getDestinationFilePath(fileName string) string
}

type S3Destination struct {
	client     *s3.Client
	bucketName string
	prefix     *string
}

func NewS3Destination(awsConfig aws.Config, destinationWithoutPrefix string) (S3Destination, error) {
	s3Client := s3.NewFromConfig(awsConfig)

	bucketName, prefix, err := parseS3Destination(destinationWithoutPrefix)
	if err != nil {
		return S3Destination{}, err
	}

	return S3Destination{
		client:     s3Client,
		bucketName: bucketName,
		prefix:     prefix,
	}, nil
}

func (d *S3Destination) Display() string {
	output := fmt.Sprintf("arn:aws:s3:::%s", d.bucketName)

	if d.prefix != nil {
		output = fmt.Sprintf("%s/%s", output, *d.prefix)
	}

	return output
}

func (d *S3Destination) getDestinationFilePath(fileName string) string {
	filePath := fileName
	if d.prefix != nil {
		filePath = fmt.Sprintf("%s/%s", *d.prefix, fileName)
	}

	return filePath
}

func (d *S3Destination) upload(ctx context.Context, body io.Reader, fileName string) error {
	objectKey := d.getDestinationFilePath(fileName)

	_, err := d.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(d.bucketName),
		Key:         aws.String(objectKey),
		Body:        body,
		ContentType: aws.String("text/plain"),
	})

	return err
}

func parseS3Destination(destinationWithoutPrefix string) (string, *string, error) {
	destinationWithoutPrefix = strings.TrimSpace(destinationWithoutPrefix)

	if destinationWithoutPrefix == "" {
		return "", nil, fmt.Errorf("destination is empty")
	}

	parts := strings.SplitN(destinationWithoutPrefix, "/", 2)
	bucketName := parts[0]

	if bucketName == "" {
		return "", nil, fmt.Errorf("invalid S3 ARN: bucket name is empty")
	}

	var prefix *string
	if len(parts) > 1 && parts[1] != "" {
		prefix = &parts[1]
	}

	return bucketName, prefix, nil
}
