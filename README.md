# kplay

✨ Overview
---

`kplay` lets you inspect messages in a Kafka topic in a simple and deliberate
manner. Using it, you can pull one or more records on demand, peruse through
them in a list, and, if needed, persist them to your local filesystem.

<p align="center">
  <img src="./assets/kplay.gif?raw=true" alt="Usage" />
</p>

Install
---

**homebrew**:

```sh
brew install dhth/tap/kplay
```

**go**:

```sh
go install github.com/dhth/kplay@latest
```

⚡️ Usage
---

### Consuming JSON messages

As a binary, kplay only supports consuming JSON messages.

```bash
kplay \
    -brokers='<COMMA_SEPARATED_BROKER_URLS>' \
    -topic='<TOPIC>' \
    -group='<CONSUMER-GROUP>'
```

### Consuming protobuf messages

Protobuf messages can be consumed, but that will need some tweaks to the source
code.

Place your protobuf files under `./proto`. Using the protoc compiler, run:

```bash
protoc --go_out=. proto/<YOUR_PROTOBUF_FILE>.proto
```

Change the generated struct reference in `./ui/model/utils.go`.

Compile, and run.

### Authentication

By default, `kplay` operates under the assumption that brokers do not
authenticate requests. Besides this, it also supports [AWS IAM
authentication][2] (turned on via `-auth='msk_iam_auth'`).

TODO
---

- [ ] Add ability to only save records that match a chosen set of keys

Acknowledgements
---

`kplay` is built using the awesome TUI framework [bubbletea][1].

[1]: https://github.com/charmbracelet/bubbletea
[2]: https://docs.aws.amazon.com/msk/latest/developerguide/iam-access-control.html
