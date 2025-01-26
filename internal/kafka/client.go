package kafka

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	c "github.com/dhth/kplay/internal/config"
	"github.com/twmb/franz-go/pkg/kgo"
	kaws "github.com/twmb/franz-go/pkg/sasl/aws"
)

var (
	errCouldntRetrieveAWSCredentials = errors.New("couldn't retrieve AWS credentials")
	errCouldntCreateAWSSession       = errors.New("couldn't create AWS session")
)

func GetClient(auth c.AuthType, brokers []string, group, topic string) (*kgo.Client, error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup(group),
		kgo.ConsumeTopics(topic),
		kgo.DisableAutoCommit(),
	}
	if auth == c.AWSMSKIAM {
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
		opts = append(opts, kgo.Dialer((&tls.Dialer{NetDialer: &net.Dialer{Timeout: 10 * time.Second}}).DialContext))
	}

	return kgo.NewClient(opts...)
}
