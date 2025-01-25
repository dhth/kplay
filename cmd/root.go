package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"strings"
	"time"

	d "github.com/dhth/kplay/internal/domain"
	k "github.com/dhth/kplay/internal/kafka"
	"github.com/dhth/kplay/internal/ui"
)

var (
	errBrokersEmpty             = errors.New("brokers cannot be empty")
	errTopicEmpty               = errors.New("topic cannot be empty")
	errGroupEmpty               = errors.New("group cannot be empty")
	errAuthEmpty                = errors.New("auth cannot be empty")
	errUnsupportedAuth          = errors.New("unsupported auth provided")
	errCouldntCreateKafkaClient = errors.New("couldn't create kafka client")
	errCouldntPingBrokers       = errors.New("couldn't ping the brokers")
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

	var authType k.AuthType
	switch *auth {
	case "none":
		authType = k.NoAuth
	case "msk_iam_auth":
		authType = k.SaslIamAuth
	default:
		return fmt.Errorf("%w; supported values: none, sasl_iam_auth", errUnsupportedAuth)
	}

	cl, err := k.GetClient(authType, strings.Split(*brokers, ","), *group, *topic)
	if err != nil {
		return fmt.Errorf("%w: %s", errCouldntCreateKafkaClient, err.Error())
	}

	defer cl.Close()

	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	if err := cl.Ping(ctx); err != nil {
		return fmt.Errorf("%w: %s", errCouldntPingBrokers, err.Error())
	}

	return ui.RenderUI(cl, kconfig)
}
