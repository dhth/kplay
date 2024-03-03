package cmd

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dhth/kplay/ui"
	"github.com/twmb/franz-go/pkg/kgo"
)

func die(msg string, args ...any) {
	fmt.Fprintf(os.Stderr, msg, args...)
	os.Exit(1)
}

var (
	seedBrokers = flag.String("brokers", "127.0.0.1:9092", "comma delimited list of seed brokers")
	topic       = flag.String("topic", "", "topic to consume from")
	style       = flag.String("commit-style", "autocommit", "commit style (which consume & commit is chosen); autocommit|records|uncommitted")
	group       = flag.String("group", "", "group to consume within")
	logger      = flag.Bool("logger", false, "if true, enable an info level logger")
)

func Execute() {
	flag.Parse()

	styleNum := 0
	switch {
	case strings.HasPrefix("autocommit", *style):
	case strings.HasPrefix("records", *style):
		styleNum = 1
	case strings.HasPrefix("uncommitted", *style):
		styleNum = 2
	default:
		die("unrecognized style %s", *style)
	}

	opts := []kgo.Opt{
		kgo.SeedBrokers(strings.Split(*seedBrokers, ",")...),
		kgo.ConsumerGroup(*group),
		kgo.ConsumeTopics(*topic),
	}
	if styleNum != 0 {
		opts = append(opts, kgo.DisableAutoCommit())
	}
	if *logger {
		opts = append(opts, kgo.WithLogger(kgo.BasicLogger(os.Stderr, kgo.LogLevelInfo, nil)))
	}

	cl, err := kgo.NewClient(opts...)
	if err != nil {
		die("unable to create client: %v", err)
	}

	defer cl.Close()
	defer fmt.Println("Closed connection to broker")

	ui.RenderUI(cl)

}
