package kafka

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	t "github.com/dhth/kplay/internal/types"
	"github.com/twmb/franz-go/pkg/kgo"
	kaws "github.com/twmb/franz-go/pkg/sasl/aws"
)

var (
	errCouldntRetrieveAWSCredentials = errors.New("couldn't retrieve AWS credentials")
	errCouldntLoadAwsConfig          = errors.New("couldn't load AWS config")
	errCouldntCreateKafkaClient      = errors.New("couldn't create kafka client")
)

type Builder struct {
	opts []kgo.Opt
}

func NewBuilder(brokers []string, topic string) Builder {
	opts := []kgo.Opt{
		kgo.SeedBrokers(brokers...),
		kgo.ConsumeTopics(topic),
	}

	return Builder{opts}
}

func (b Builder) WithMskIAMAuth(awsCfg aws.Config) Builder {
	authFn := func(c context.Context) (kaws.Auth, error) {
		creds, err := awsCfg.Credentials.Retrieve(c)
		if err != nil {
			return kaws.Auth{}, fmt.Errorf("%w: %w", errCouldntRetrieveAWSCredentials, err)
		}

		return kaws.Auth{
			AccessKey:    creds.AccessKeyID,
			SecretKey:    creds.SecretAccessKey,
			SessionToken: creds.SessionToken,
			UserAgent:    "kplay",
		}, nil
	}

	b.opts = append(b.opts, kgo.SASL(kaws.ManagedStreamingIAM(authFn)))

	dialer := tls.Dialer{
		NetDialer: &net.Dialer{
			Timeout: 10 * time.Second,
		},
	}
	b.opts = append(b.opts, kgo.Dialer(
		(&dialer).DialContext))

	return b
}

func (b Builder) WithStartOffset(offset int64) Builder {
	b.opts = append(b.opts, kgo.ConsumeStartOffset(kgo.NewOffset().At(offset)))

	return b
}

func (b Builder) WithStartTimestamp(timestamp time.Time) Builder {
	millis := timestamp.UnixMilli()
	b.opts = append(b.opts, kgo.ConsumeStartOffset(kgo.NewOffset().AfterMilli(millis)))

	return b
}

func GetKafkaClient(
	auth t.AuthType,
	brokers []string,
	topic string,
	consumeBehaviours t.ConsumeBehaviours,
) (*kgo.Client, error) {
	builder := NewBuilder(brokers, topic)

	if auth == t.AWSMSKIAM {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", errCouldntLoadAwsConfig, err.Error())
		}
		builder = builder.WithMskIAMAuth(cfg)
	}

	if consumeBehaviours.StartOffset != nil {
		builder = builder.WithStartOffset(*consumeBehaviours.StartOffset)
	} else if consumeBehaviours.StartTimeStamp != nil {
		builder = builder.WithStartTimestamp(*consumeBehaviours.StartTimeStamp)
	}

	client, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errCouldntCreateKafkaClient, err.Error())
	}

	return client, nil
}

func (b Builder) Build() (*kgo.Client, error) {
	return kgo.NewClient(b.opts...)
}
