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
		return zero, fmt.Errorf("%w: %s", t.ErrCouldntLoadAwsConfig, err.Error())
	}

	return cfg, nil
}
