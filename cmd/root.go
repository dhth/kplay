package cmd

import (
	"context"
	"crypto/tls"
	"net"

	"flag"
	"fmt"

	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dhth/kplay/ui"
	"github.com/dhth/kplay/ui/model"
	"github.com/twmb/franz-go/pkg/kgo"
	kaws "github.com/twmb/franz-go/pkg/sasl/aws"
)

func die(msg string, args ...any) {
	fmt.Fprintf(os.Stderr, msg, args...)
	os.Exit(1)
}

var (
	brokers = flag.String("brokers", "127.0.0.1:9092", "comma delimited list of brokers")
	topic   = flag.String("topic", "", "topic to consume from")
	group   = flag.String("group", "", "group to consume within")
	auth    = flag.String("auth", "none", "authentication used by the brokers")
)

func Execute() {

	if len(os.Getenv("DEBUG")) > 0 {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		defer f.Close()
	}

	flag.Parse()

	if *brokers == "" {
		die("brokers cannot be empty")
	}
	if *topic == "" {
		die("topic cannot be empty")
	}
	if *group == "" {
		die("group cannot be empty")
	}
	if *auth == "" {
		die("auth cannot be empty")
	}

	deserFmt := model.JsonFmt

	kconfig := model.KConfig{
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
		die("unsupported authentication type; supported values: none, sasl_iam_auth")
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		die("error creating AWS session: %v", err)
	}

	opts := []kgo.Opt{
		kgo.SeedBrokers(strings.Split(*brokers, ",")...),
		kgo.ConsumerGroup(*group),
		kgo.ConsumeTopics(*topic),
		kgo.DisableAutoCommit(),
	}
	switch authType {
	case SaslIamAuth:
		opts = append(opts, kgo.SASL(kaws.ManagedStreamingIAM(func(ctx context.Context) (kaws.Auth, error) {
			creds, err := cfg.Credentials.Retrieve(context.TODO())
			if err != nil {
				return kaws.Auth{}, err
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
		die("unable to create client: %v\n", err)
	}

	defer cl.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := cl.Ping(ctx); err != nil {
		cl.Close()
		die("cannot ping the broker(s): %s\n", err)
	}

	ui.RenderUI(cl, kconfig)

}
