package cmd

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"

	d "github.com/dhth/kplay/internal/domain"
	"github.com/dhth/kplay/internal/ui"
	"github.com/twmb/franz-go/pkg/kgo"
	kaws "github.com/twmb/franz-go/pkg/sasl/aws"
)

var (
	errBrokersEmpty                  = errors.New("brokers cannot be empty")
	errTopicEmpty                    = errors.New("topic cannot be empty")
	errGroupEmpty                    = errors.New("group cannot be empty")
	errAuthEmpty                     = errors.New("auth cannot be empty")
	errUnsupportedAuth               = errors.New("unsupported auth provided")
	errCouldntCreateAWSSession       = errors.New("couldn't create AWS session")
	errCouldntRetrieveAWSCredentials = errors.New("couldn't retrieve AWS credentials")
	errCouldntCreateKafkaClient      = errors.New("couldn't create kafka client")
	errCouldntPingBrokers            = errors.New("couldn't ping the brokers")
)

var (
	brokers = flag.String("brokers", "127.0.0.1:9092", "comma delimited list of brokers")
	topic   = flag.String("topic", "", "topic to consume from")
	group   = flag.String("group", "", "group to consume within")
	auth    = flag.String("auth", "none", "authentication used by the brokers")
)

func Execute() error {
	flag.Parse()

	if strings.TrimSpace(*brokers) == "" {
		return errBrokersEmpty
	}
	if strings.TrimSpace(*topic) == "" {
		return errTopicEmpty
	}
	if strings.TrimSpace(*group) == "" {
		return errGroupEmpty
	}
	if strings.TrimSpace(*auth) == "" {
		return errAuthEmpty
	}

	deserFmt := d.JSON

	kconfig := d.Config{
		Topic:         *topic,
		ConsumerGroup: *group,
		DeserFmt:      deserFmt,
	}

	var authType KafkaAuthenticationType
	switch *auth {
	case "none":
		authType = NoAuth
	case "msk_iam_auth":
		authType = SaslIamAuth
	default:
		return fmt.Errorf("%w; supported values: none, sasl_iam_auth", errUnsupportedAuth)
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return fmt.Errorf("%w: %s", errCouldntCreateAWSSession, err.Error())
	}

	opts := []kgo.Opt{
		kgo.SeedBrokers(strings.Split(*brokers, ",")...),
		kgo.ConsumerGroup(*group),
		kgo.ConsumeTopics(*topic),
		kgo.DisableAutoCommit(),
	}
	if authType == SaslIamAuth {
		opts = append(opts, kgo.SASL(kaws.ManagedStreamingIAM(func(_ context.Context) (kaws.Auth, error) {
			creds, err := cfg.Credentials.Retrieve(context.TODO())
			if err != nil {
				return kaws.Auth{}, fmt.Errorf("%w: %w", errCouldntRetrieveAWSCredentials, err)
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

	cl, err := kgo.NewClient(opts...)
	if err != nil {
		return fmt.Errorf("%w: %s", errCouldntCreateKafkaClient, err.Error())
	}

	defer cl.Close()

	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()
	if err := cl.Ping(ctx); err != nil {
		cl.Close()
		return fmt.Errorf("%w: %s", errCouldntPingBrokers, err.Error())
	}

	return ui.RenderUI(cl, kconfig)
}
