# kplay

[![Build Workflow Status](https://img.shields.io/github/actions/workflow/status/dhth/kplay/main.yml?style=flat-square)](https://github.com/dhth/kplay/actions/workflows/main.yml)
[![Vulncheck Workflow Status](https://img.shields.io/github/actions/workflow/status/dhth/kplay/vulncheck.yml?style=flat-square&label=vulncheck)](https://github.com/dhth/kplay/actions/workflows/vulncheck.yml)
[![Latest Release](https://img.shields.io/github/release/dhth/kplay.svg?style=flat-square)](https://github.com/dhth/kplay/releases/latest)
[![Commits Since Latest Release](https://img.shields.io/github/commits-since/dhth/kplay/latest?style=flat-square)](https://github.com/dhth/kplay/releases)

`kplay` (short for "kafka-playground") lets you inspect messages in a Kafka
topic in a simple and deliberate manner. Using it, you can pull one or more
messages on demand, decode them based on a configured encoding format, peruse
them in a list, persist them to your local filesystem, or forward them to S3.

<video src="https://github.com/user-attachments/assets/c06ec742-06da-4836-ac33-ef25d3a40786"></video>

![tui](https://github.com/user-attachments/assets/613727e7-bca8-4855-b19c-bed2faf80314)

![web](https://github.com/user-attachments/assets/e3af71a2-8f06-4b9b-8e48-f96ad2c0f972)

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

Or get the binaries directly from a [release][4]. Read more about verifying the
authenticity of released artifacts [here](#-verifying-release-artifacts).

⚡️ Usage
---

`kplay` offers 4 commands:

- `tui`: browse messages in a kafka topic via a TUI
- `serve`: browse messages in a kafka topic via a web interface
- `scan`: scan a topic for messages, and optionally save them to your local
    filesystem
- `forward`: consume messages from a topic, and forward them to a remote
    destination

### TUI

This will start a TUI which will let you browse messages on demand. You
can then browse the message metadata and value in a pager. By default, kplay
will consume messages from the earliest possible offset, but you can modify this
behaviour by either providing an offset or a timestamp to start consuming
messages from.

```text
Usage:
  kplay tui <PROFILE> [flags]

Flags:
  -o, --from-offset string      start consuming messages from this offset; provide a single offset for all partitions (eg. 1000) or specify offsets per partition (e.g., '0:1000,2:1500')
  -t, --from-timestamp string   start consuming messages from this timestamp (in RFC3339 format, e.g., 2006-01-02T15:04:05Z07:00)
  -h, --help                    help for tui
  -O, --output-dir string       directory to persist messages in (default "$HOME/.kplay")
  -p, --persist-messages        whether to start the TUI with the setting "persist messages" ON
  -s, --skip-messages           whether to start the TUI with the setting "skip messages" ON

Global Flags:
  -c, --config-path string   location of kplay's config file (can also be provided via $KPLAY_CONFIG_PATH)
      --debug                whether to only display config picked up by kplay without running it
```

[![tui](https://asciinema.org/a/vHSXtmOfIyh5DaRE5SlGzdmJ8.svg)](https://asciinema.org/a/vHSXtmOfIyh5DaRE5SlGzdmJ8)

#### ⌨️ TUI Keymaps

### General

| Keymap        | Action           |
|---------------|------------------|
| `?`           | Show help view   |
| `q` / `<esc>` | Go back/quit     |
| `<ctrl+c>`    | Quit immediately |

### Message List and Details View

| Keymap                  | Action                                         |
|-------------------------|------------------------------------------------|
| `<tab>` / `<shift-tab>` | Switch focus between panes                     |
| `j` / `<Down>`          | Select next message / scroll details down      |
| `k` / `<Up>`            | Select previous message / scroll details up    |
| `G`                     | Select last message / scroll details to bottom |
| `g`                     | Select first message / scroll details to top   |
| `<ctrl+d>`              | Scroll details half page down                  |
| `<ctrl+u>`              | Scroll details half page up                    |
| `]`                     | Select next message                            |
| `[`                     | Select previous message                        |
| `n`                     | Fetch the next message from the topic          |
| `N`                     | Fetch the next 10 messages from the topic      |
| `}`                     | Fetch the next 100 messages from the topic     |
| `s`                     | Toggle skipping mode                           |
| `p`                     | Toggle persist mode                            |
| `P`                     | Persist current message to local filesystem    |
| `y`                     | Copy message details to clipboard              |

### Serve

This will start `kplay`'s web interface which will let you browse messages on
demand. By default, kplay will consume messages from the earliest possible
offset, but you can modify this behaviour by either providing an offset or a
timestamp to start consuming messages from.

```text
Usage:
  kplay serve <PROFILE> [flags]

Flags:
  -o, --from-offset string      start consuming messages from this offset; provide a single offset for all partitions (eg. 1000) or specify offsets per partition (e.g., '0:1000,2:1500')
  -t, --from-timestamp string   start consuming messages from this timestamp (in RFC3339 format, e.g., 2006-01-02T15:04:05Z07:00)
  -h, --help                    help for serve
  -O, --open                    whether to open web interface in browser automatically
  -S, --select-on-hover         whether to start the web interface with the setting "select on hover" ON

Global Flags:
  -c, --config-path string   location of kplay's config file (can also be provided via $KPLAY_CONFIG_PATH)
      --debug                whether to only display config picked up by kplay without running it
```

![web](https://tools.dhruvs.space/images/kplay/kplay-web.gif)

### Scan

This command is useful when you want to view a summary of messages in a Kafka
topic (ie, the partition, offset, timestamp, and key of each message), and
optionally save the message values to your local filesystem.

```text
Usage:
  kplay scan <PROFILE> [flags]

Flags:
  -b, --batch-size uint         number of messages to fetch per batch (must be greater than 0) (default 100)
  -d, --decode                  whether to decode message values (false is equivalent to 'encodingFormat: raw' in kplay's config) (default true)
  -o, --from-offset string      scan messages from this offset; provide a single offset for all partitions (eg. 1000) or specify offsets per partition (e.g., '0:1000,2:1500')
  -t, --from-timestamp string   scan messages from this timestamp (in RFC3339 format, e.g., 2006-01-02T15:04:05Z07:00)
  -h, --help                    help for scan
  -k, --key-regex string        regex to filter message keys by
  -n, --num-records uint        maximum number of messages to scan (default 1000)
  -O, --output-dir string       directory to save scan results in (default "$HOME/.kplay")
  -s, --save-messages           whether to save kafka messages to the local filesystem

Global Flags:
  -c, --config-path string   location of kplay's config file (can also be provided via $KPLAY_CONFIG_PATH)
      --debug                whether to only display config picked up by kplay without running it
```

[![scan](https://asciinema.org/a/NutRtcDkmtYLLTCZ3eVe4CfNx.svg)](https://asciinema.org/a/NutRtcDkmtYLLTCZ3eVe4CfNx)

### Forward

This command is useful when you want to consume messages in a kafka topic as
part of a consumer group, decode them, and forward the decoded contents to a
remote destination (AWS S3 is the only supported destination for now).

This command is intended to be run in a long running containerised environment;
as such, it accepts configuration via the following environment variables.

| Environment Variable                    | Description                                           | Default Value   | Range       |
|-----------------------------------------|-------------------------------------------------------|-----------------|-------------|
| KPLAY_FORWARD_CONSUMER_GROUP            | Consumer group to use                                 | kplay-forwarder | -           |
| KPLAY_FORWARD_FETCH_BATCH_SIZE          | Number of records to fetch per batch                  | 50              | 1-1000      |
| KPLAY_FORWARD_NUM_UPLOAD_WORKERS        | Number of upload workers                              | 50              | 1-500       |
| KPLAY_FORWARD_SHUTDOWN_TIMEOUT_MILLIS   | Graceful shutdown timeout in ms                       | 30000           | 10000-60000 |
| KPLAY_FORWARD_POLL_FETCH_TIMEOUT_MILLIS | Kafka polling fetch timeout in ms                     | 10000           | 1000-60000  |
| KPLAY_FORWARD_POLL_SLEEP_MILLIS         | Kafka polling sleep interval in ms                    | 5000            | 0-1800000   |
| KPLAY_FORWARD_UPLOAD_TIMEOUT_MILLIS     | Upload timeout in ms                                  | 10000           | 1000-60000  |
| KPLAY_FORWARD_UPLOAD_REPORTS            | Whether to upload reports of the messages forwarded   | false           | -           |
| KPLAY_FORWARD_REPORT_BATCH_SIZE         | Report batch size                                     | 5000            | 1000-20000  |
| KPLAY_FORWARD_RUN_SERVER                | Whether to run an HTTP server alongside the forwarder | false           | -           |
| KPLAY_FORWARD_SERVER_HOST               | Host to run the server on                             | 127.0.0.1       | -           |
| KPLAY_FORWARD_SERVER_PORT               | Port to run the server on                             | 8080            | -           |

If needed, this command can also start an HTTP server which can be used for
health checks (at `/health`).

```text
Usage:
  kplay forward <PROFILE>,<PROFILE>,... <DESTINATION> [flags]

Examples:
kplay forward profile-1,profile-2 arn:aws:s3:::bucket-to-forward-messages-to/prefix

Flags:
  -h, --help   help for forward

Global Flags:
  -c, --config-path string   location of kplay's config file (can also be provided via $KPLAY_CONFIG_PATH)
      --debug                whether to only display config picked up by kplay without running it
```

[![forward](https://asciinema.org/a/ivVUXTSfkacmRPFNIUmUSnDkX.svg)](https://asciinema.org/a/ivVUXTSfkacmRPFNIUmUSnDkX)

🔧 Configuration
---

kplay's configuration file looks like the following:

```yaml
profiles:
  - name: json-encoded
    authentication: none
    encodingFormat: json
    brokers:
      - 127.0.0.1:9092
    topic: kplay-test-1

  - name: proto-encoded
    authentication: aws_msk_iam
    encodingFormat: protobuf
    protoConfig:
      descriptorSetFile: path/to/descriptor/set/file.pb
      descriptorName: sample.DescriptorName
    brokers:
      - 127.0.0.1:9092
    topic: kplay-test-2

  - name: raw
    authentication: none
    encodingFormat: raw
    brokers:
      - 127.0.0.1:9092
    topic: kplay-test-3

  - name: tls-enabled
    authentication: none
    encodingFormat: json
    tlsConfig:
      enabled: true
      insecureSkipVerify: false
    brokers:
      - kafka.example.com:9093
    topic: kplay-test-tls
```

### TLS/SSL Configuration

To enable TLS/SSL for secure communication with Kafka brokers, add a `tlsConfig` section to your profile:

```yaml
tlsConfig:
  enabled: true                # Enable TLS/SSL (default: false)
  insecureSkipVerify: false    # Skip certificate verification (default: false, use true for self-signed certs in dev)
```

**Note:** When using AWS MSK IAM authentication, TLS is automatically enabled and configured.

🔤 Message Encoding
---

`kplay` supports decoding messages that are encoded in two data formats: JSON
and protobuf. It also supports handling the message bytes as raw data (using the
`encodingFormat` "raw").

### Decoding protobuf encoded messages

For decoding protobuf encoded messages, `kplay` needs to be provided with a
`FileDescriptorSet` and a descriptor name. Consider a .proto file like the
following:

```text
// application_state.proto
syntax = "proto3";

package sample;

message ApplicationState {
  string id = 1; // required
  string colorTheme = 2;
  string backgroundImageUrl = 3;
  string customDomain = 4;
}
```

A `FileDescriptorSet` can be generated for this file using the [protocol buffer
compiler][5].

```bash
protoc application_state.proto \
    --descriptor_set_out=application_state.pb \
    --include_imports
```

This descriptor set file can then be used in `kplay`'s config file, alongside
the `descriptorName` "sample.ApplicationState".

> Read more about self describing protocol messages [here][3].

🔑 Authentication
---

By default, `kplay` operates under the assumption that brokers do not
authenticate requests. Besides this, it supports [AWS IAM authentication][2].

🔐 Verifying release artifacts
---

In case you get the `kplay` binary directly from a [release][4], you may want to
verify its authenticity. Checksums are applied to all released artifacts, and
the resulting checksum file is signed using
[cosign](https://docs.sigstore.dev/cosign/installation/).

Steps to verify (replace `A.B.C` in the commands listed below with the version
you want):

1. Download the following files from the release:

    - kplay_A.B.C_checksums.txt
    - kplay_A.B.C_checksums.txt.pem
    - kplay_A.B.C_checksums.txt.sig

2. Verify the signature:

   ```shell
   cosign verify-blob kplay_A.B.C_checksums.txt \
       --certificate kplay_A.B.C_checksums.txt.pem \
       --signature kplay_A.B.C_checksums.txt.sig \
       --certificate-identity-regexp 'https://github\.com/dhth/kplay/\.github/workflows/.+' \
       --certificate-oidc-issuer "https://token.actions.githubusercontent.com"
   ```

3. Download the compressed archive you want, and validate its checksum:

   ```shell
   curl -sSLO https://github.com/dhth/kplay/releases/download/vA.B.C/kplay_A.B.C_linux_amd64.tar.gz
   sha256sum --ignore-missing -c kplay_A.B.C_checksums.txt
   ```

3. If checksum validation goes through, uncompress the archive:

   ```shell
   tar -xzf kplay_A.B.C_linux_amd64.tar.gz
   ./kplay
   # profit!
   ```

Acknowledgements
---

`kplay` is built using the awesome TUI framework [bubbletea][1].

[1]: https://github.com/charmbracelet/bubbletea
[2]: https://docs.aws.amazon.com/msk/latest/developerguide/iam-access-control.html
[3]: https://protobuf.dev/programming-guides/techniques/#self-description
[4]: https://github.com/dhth/kplay/releases
[5]: https://grpc.io/docs/protoc-installation
