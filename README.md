# kplay

[![Build Workflow Status](https://img.shields.io/github/actions/workflow/status/dhth/kplay/main.yml?style=flat-square)](https://github.com/dhth/kplay/actions/workflows/main.yml)
[![Vulncheck Workflow Status](https://img.shields.io/github/actions/workflow/status/dhth/kplay/vulncheck.yml?style=flat-square&label=vulncheck)](https://github.com/dhth/kplay/actions/workflows/vulncheck.yml)
[![Latest Release](https://img.shields.io/github/release/dhth/kplay.svg?style=flat-square)](https://github.com/dhth/kplay/releases/latest)
[![Commits Since Latest Release](https://img.shields.io/github/commits-since/dhth/kplay/latest?style=flat-square)](https://github.com/dhth/kplay/releases)

`kplay` (short for "kafka-playground") lets you inspect messages in a Kafka
topic in a simple and deliberate manner. Using it, you can pull one or more
messages on demand, peruse through them in a list, and, if needed, persist them
to your local filesystem.

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

`kplay` can display messages via two interfaces: a TUI or a webpage 

```text
$ kplay tui -h
open kplay's TUI

Usage:
  kplay tui <PROFILE> [flags]

Flags:
  -C, --commit-messages         whether to start the TUI with the setting "commit messages" ON (default true)
  -c, --config-path string      location of kplay's config file (default "/Users/dhruvthakur/Library/Application Support/kplay/kplay.yml")
  -g, --consumer-group string   consumer group to use (overrides the one in kplay's config file)
      --debug                   whether to only display config picked up by kplay without running it
  -h, --help                    help for tui
  -p, --persist-messages        whether to start the TUI with the setting "persist messages" ON
  -s, --skip-messages           whether to start the TUI with the setting "skip messages" ON
```

https://github.com/user-attachments/assets/e7a1aa58-21d2-45fd-827a-454445a97e6e

```text
$ kplay serve -h
open kplay's web interface

Usage:
  kplay serve <PROFILE> [flags]

Flags:
  -C, --commit-messages         whether to start the web interface with the setting "commit messages" ON (default true)
  -c, --config-path string      location of kplay's config file (default "/Users/dhruvthakur/Library/Application Support/kplay/kplay.yml")
  -g, --consumer-group string   consumer group to use (overrides the one in kplay's config file)
      --debug                   whether to only display config picked up by kplay without running it
  -h, --help                    help for serve
  -o, --open                    whether to open web interface in browser automatically
  -S, --select-on-hover         whether to start the web interface with the setting "select on hover" ON
```

https://github.com/user-attachments/assets/dc52af12-0cc4-41f1-b4f0-9904291fa721

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
    consumerGroup: kplay-consumer-group-1

  - name: proto-encoded
    authentication: aws_msk_iam
    encodingFormat: protobuf
    protoConfig:
      descriptorSetFile: path/to/descriptor/set/file.pb
      descriptorName: sample.DescriptorName
    brokers:
      - 127.0.0.1:9092
    topic: kplay-test-2
    consumerGroup: kplay-consumer-group-1

  - name: raw
    authentication: none
    encodingFormat: raw
    brokers:
      - 127.0.0.1:9092
    topic: kplay-test-3
    consumerGroup: kplay-consumer-group-1
```

🔤 Message Encoding
---

`kplay` supports parsing messages that are encoded in two data formats: JSON and
protobuf. It also supports parsing raw data (using the `encodingFormat` "raw").

### Parsing protobuf encoded messages

For parsing protobuf encoded messages, `kplay` needs to be provided with a
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

⌨️ TUI Keymaps
---

### General

| Keymap | Description        |
|--------|--------------------|
| `?`    | Show help view     |
| `q`    | Go back/quit       |
| `Q`    | Quit from anywhere |

### Message List and Details View

| Keymap                | Description                                                                                                                                  |
|-----------------------|----------------------------------------------------------------------------------------------------------------------------------------------|
| `<tab>`/`<shift-tab>` | Switch focus between panes                                                                                                                   |
| `j`/`<Down>`          | Move cursor/details pane down                                                                                                                |
| `k`/`<Up>`            | Move cursor/details pane up                                                                                                                  |
| `n`                   | Fetch the next message from the topic                                                                                                        |
| `N`                   | Fetch the next 10 messages from the topic                                                                                                    |
| `}`                   | Fetch the next 100 messages from the topic                                                                                                   |
| `s`                   | Toggle skipping mode (if ON, kplay will consume messages, but not populate its internal list, effectively skipping over them)                |
| `p`                   | Toggle persist mode (if ON, kplay will start persisting messages at the location messages/<topic>/partition-<partition>/offset-<offset>.txt) |
| `c`                   | Toggle commit mode (if OFF, kplay will consume messages without committing them)                                                             |
| `y`                   | Copy message details to clipboard                                                                                                            |
| `[`                   | Move to previous item in list                                                                                                                |
| `]`                   | Move to next item in list                                                                                                                    |

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
