package awsweb

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	t "github.com/dhth/kplay/internal/types"
)

func GetAWSConfig(ctx context.Context) (aws.Config, error) {
	var zero aws.Config

	configCtx, configCancel := context.WithTimeout(ctx, 5*time.Second)
	defer configCancel()

	cfg, err := config.LoadDefaultConfig(configCtx)
	if err != nil {
		return zero, fmt.Errorf("%w: %w", t.ErrCouldntLoadAwsConfig, err)
	}

	credsCtx, credsCancel := context.WithTimeout(ctx, 3*time.Second)
	defer credsCancel()
	_, err = cfg.Credentials.Retrieve(credsCtx)
	if err != nil {
		return zero, fmt.Errorf("%w: %w", t.ErrCouldntRetrieveAWSCredentials, err)
	}

	return cfg, nil
}
