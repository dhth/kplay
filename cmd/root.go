package cmd

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dhth/kplay/ui"
	"github.com/dhth/kplay/ui/model"
	"github.com/twmb/franz-go/pkg/kgo"
)

func die(msg string, args ...any) {
	fmt.Fprintf(os.Stderr, msg, args...)
	os.Exit(1)
}

var (
	brokers = flag.String("brokers", "127.0.0.1:9092", "comma delimited list of brokers")
	topic   = flag.String("topic", "", "topic to consume from")
	group   = flag.String("group", "", "group to consume within")
)

func Execute() {
	flag.Parse()

	deserFmt := model.JsonFmt

	opts := []kgo.Opt{
		kgo.SeedBrokers(strings.Split(*brokers, ",")...),
		kgo.ConsumerGroup(*group),
		kgo.ConsumeTopics(*topic),
		kgo.DisableAutoCommit(),
	}

	cl, err := kgo.NewClient(opts...)
	if err != nil {
		die("unable to create client: %v", err)
	}

	defer cl.Close()

	ui.RenderUI(cl, deserFmt)

}
