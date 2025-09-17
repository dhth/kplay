package kafka

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	t "github.com/dhth/kplay/internal/types"
	"github.com/twmb/franz-go/pkg/kgo"
	kaws "github.com/twmb/franz-go/pkg/sasl/aws"
)

var errCouldntCreateKafkaClient = errors.New("couldn't create kafka client")

type Builder struct {
	opts []kgo.Opt
}

func NewBuilder(brokers []string) Builder {
	opts := []kgo.Opt{
		kgo.SeedBrokers(brokers...),
	}

	return Builder{opts}
}

func (b Builder) WithTopic(topic string) Builder {
	b.opts = append(b.opts, kgo.ConsumeTopics(strings.Split(topic, ",")...))

	return b
}

func (b Builder) WithMskIAMAuth(awsCfg aws.Config) Builder {
	authFn := func(c context.Context) (kaws.Auth, error) {
		creds, err := awsCfg.Credentials.Retrieve(c)
		if err != nil {
			return kaws.Auth{}, fmt.Errorf("%w: %w", t.ErrCouldntRetrieveAWSCredentials, err)
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

func (b Builder) WithStartOffset(topic string, offset int64) Builder {
	b.opts = append(b.opts, kgo.ConsumeTopics(strings.Split(topic, ",")...))
	b.opts = append(b.opts, kgo.ConsumeStartOffset(kgo.NewOffset().At(offset)))

	return b
}

func (b Builder) WithPartitionOffsets(topic string, partitionOffsets map[int32]int64) Builder {
	partitions := make(map[string]map[int32]kgo.Offset)
	topicPartitions := make(map[int32]kgo.Offset)

	for partition, offset := range partitionOffsets {
		topicPartitions[partition] = kgo.NewOffset().At(offset)
	}

	partitions[topic] = topicPartitions

	b.opts = append(b.opts, kgo.ConsumePartitions(partitions))

	return b
}

func (b Builder) WithStartTimestamp(topic string, timestamp time.Time) Builder {
	millis := timestamp.UnixMilli()
	b.opts = append(b.opts, kgo.ConsumeTopics(strings.Split(topic, ",")...))
	b.opts = append(b.opts, kgo.ConsumeStartOffset(kgo.NewOffset().AfterMilli(millis)))

	return b
}

func (b Builder) WithConsumerGroup(topic, group string) Builder {
	b.opts = append(b.opts, kgo.ConsumeTopics(strings.Split(topic, ",")...))
	b.opts = append(b.opts, kgo.ConsumerGroup(group))

	return b
}

func GetKafkaClient(
	auth t.AuthType,
	brokers []string,
	topic string,
	consumeBehaviours t.ConsumeBehaviours,
	awsCfg *aws.Config,
) (*kgo.Client, error) {
	builder := NewBuilder(brokers)

	if auth == t.AWSMSKIAM {
		builder = builder.WithMskIAMAuth(*awsCfg)
	}

	if consumeBehaviours.StartTimeStamp != nil {
		builder = builder.WithStartTimestamp(topic, *consumeBehaviours.StartTimeStamp)
	} else if consumeBehaviours.StartOffset != nil {
		builder = builder.WithStartOffset(topic, *consumeBehaviours.StartOffset)
	} else if len(consumeBehaviours.PartitionOffsets) > 0 {
		builder = builder.WithPartitionOffsets(topic, consumeBehaviours.PartitionOffsets)
	} else {
		builder = builder.WithTopic(topic)
	}

	client, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errCouldntCreateKafkaClient, err.Error())
	}

	return client, nil
}

func GetKafkaClientForForwarding(
	auth t.AuthType,
	brokers []string,
	topic string,
	consumerGroup string,
	awsCfg *aws.Config,
) (*kgo.Client, error) {
	builder := NewBuilder(brokers)

	if auth == t.AWSMSKIAM {
		builder = builder.WithMskIAMAuth(*awsCfg)
	}

	builder = builder.WithConsumerGroup(topic, consumerGroup)

	client, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errCouldntCreateKafkaClient, err.Error())
	}

	return client, nil
}

func (b Builder) Build() (*kgo.Client, error) {
	return kgo.NewClient(b.opts...)
}
