package kafka

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	t "github.com/dhth/kplay/internal/types"
	"github.com/twmb/franz-go/pkg/kgo"
	kaws "github.com/twmb/franz-go/pkg/sasl/aws"
)

var (
	errCouldntRetrieveAWSCredentials = errors.New("couldn't retrieve AWS credentials")
	errCouldntCreateAWSSession       = errors.New("couldn't create AWS session")
)

func GetClient(auth t.AuthType, brokers []string, group, topic string) (*kgo.Client, error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup(group),
		kgo.ConsumeTopics(topic),
		kgo.DisableAutoCommit(),
	}
	if auth == t.AWSMSKIAM {
		cfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("%w: %s", errCouldntCreateAWSSession, err.Error())
		}

		opts = append(opts, kgo.SASL(kaws.ManagedStreamingIAM(func(c context.Context) (kaws.Auth, error) {
			zero := kaws.Auth{}
			creds, err := cfg.Credentials.Retrieve(c)
			if err != nil {
				return zero, fmt.Errorf("%w: %w", errCouldntRetrieveAWSCredentials, err)
			}
			return kaws.Auth{
				AccessKey:    creds.AccessKeyID,
				SecretKey:    creds.SecretAccessKey,
				SessionToken: creds.SessionToken,
				UserAgent:    "kplay",
			}, nil
		})),
		)
		dialer := tls.Dialer{
			NetDialer: &net.Dialer{
				Timeout: 10 * time.Second,
			},
		}
		opts = append(opts, kgo.Dialer(
			(&dialer).DialContext))
	}

	return kgo.NewClient(opts...)
}
