# kplay

[![Build Workflow Status](https://img.shields.io/github/actions/workflow/status/dhth/kplay/build.yml?style=flat-square)](https://github.com/dhth/kplay/actions/workflows/build.yml)
[![Vulncheck Workflow Status](https://img.shields.io/github/actions/workflow/status/dhth/kplay/vulncheck.yml?style=flat-square&label=vulncheck)](https://github.com/dhth/kplay/actions/workflows/vulncheck.yml)
[![Latest Release](https://img.shields.io/github/release/dhth/kplay.svg?style=flat-square)](https://github.com/dhth/kplay/releases/latest)
[![Commits Since Latest Release](https://img.shields.io/github/commits-since/dhth/kplay/latest?style=flat-square)](https://github.com/dhth/kplay/releases)

`kplay` (short for "kafka-playground") lets you inspect messages in a Kafka
topic in a simple and deliberate manner. Using it, you can pull one or more
messages on demand, peruse through them in a list, and, if needed, persist them
to your local filesystem.

![demo](https://github.com/user-attachments/assets/e64e148c-f267-4393-9f35-e563045ab765)

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

```text
# kplay -h

kplay ("kafka playground") lets you inspect messages in a Kafka topic in a simple and deliberate manner.

kplay relies on a configuration file that contains profiles for various Kafka topics, each with its own details related
to brokers, message encoding, authentication, etc.

Usage:
  kplay <PROFILE> [flags]

Flags:
  -c, --config-path string      location of kplay's config file (default "/Users/dhruvthakur/Library/Application Support/kplay/kplay.yml")
  -g, --consumer-group string   consumer group to use (overrides the one in kplay's config file)
      --display-config-only     whether to only display config picked up by kplay
  -h, --help                    help for kplay
  -p, --persist-messages        whether to start the TUI with the "persist messages" setting ON
  -s, --skip-messages           whether to start the TUI with the "skip messages" setting ON
```

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

kplay supports parsing messages that are encoded in two data formats: JSON and
protobuf. It also supports parsing raw data (using the `encodingFormat` "raw").

### Parsing protobuf encoded messages

For parsing protobuf encoded messages, `kplay` needs to be provided with a
`FileDescriptorSet` and a descriptor name. For example, consider a proto file
like the following:

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

A `FileDescriptorSet` can be generated for this file using the protocol buffer
compiler.

```bash
protoc application_state.proto --descriptor_set_out=application_state.pb --include_imports
```

This descriptor set file can then be used in `kplay`'s config file, alongside
the `descriptorName` "sample.ApplicationState".

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
