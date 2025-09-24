package serde

import (
	"bytes"
	_ "embed"
	"os"
	"os/exec"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/values/a.bin
var sampleMsg []byte

func TestDecodeRawProtocCompatibility(t *testing.T) {
	if os.Getenv("KPLAY_PROTOC_COMPATIBILITY_TEST") != "1" {
		t.Skip("Skipping protoc compatibility test. Set KPLAY_PROTOC_COMPATIBILITY_TEST=1 to run.")
	}

	if _, err := exec.LookPath("protoc"); err != nil {
		t.Fatalf("protoc not found in PATH")
	}

	testCases := []struct {
		name string
		data []byte
	}{
		{
			name: "msg-a",
			data: sampleMsg,
		},
		{
			name: "single_varint",
			data: []byte{0x08, 0x96, 0x01}, // field 1: varint 150
		},
		{
			name: "single_string",
			data: []byte{0x12, 0x04, 0x74, 0x65, 0x73, 0x74}, // field 2: string "test"
		},
		{
			name: "multiple_fields",
			data: []byte{
				0x08, 0x96, 0x01, // field 1: varint 150
				0x12, 0x04, 0x74, 0x65, 0x73, 0x74, // field 2: string "test"
				0x18, 0x7f, // field 3: varint 127
			},
		},
		{
			name: "nested_message",
			data: []byte{
				0x1a, 0x05, // field 3: length-delimited (5 bytes)
				0x08, 0x96, 0x01, // nested field 1: varint 150
				0x10, 0x7f, // nested field 2: varint 127
			},
		},
		{
			name: "empty_string",
			data: []byte{0x12, 0x00}, // field 2: empty string
		},
		{
			name: "large_varint",
			data: []byte{0x08, 0x80, 0x80, 0x80, 0x80, 0x08}, // field 1: large varint
		},
		{
			name: "fixed32",
			data: []byte{
				0x0d, 0x12, 0x34, 0x56, 0x78, // field 1: fixed32
			},
		},
		{
			name: "fixed64",
			data: []byte{
				0x09, 0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xde, 0xf0, // field 1: fixed64
			},
		},
		{
			name: "binary_data",
			data: []byte{
				0x12, 0x04, 0x00, 0x01, 0x02, 0x03, // field 2: binary data
			},
		},
		{
			name: "utf8_with_escapes",
			data: []byte{
				0x12, 0x09, // field 2: string (9 bytes)
				0x48, 0x65, 0x6c, 0x6c, 0x6f, // "Hello"
				0xe2, 0x80, 0x99, // UTF-8 right single quotation mark
				0x73, // "s"
			},
		},
		{
			name: "repeated_fields",
			data: []byte{
				0x08, 0x01, // field 1: varint 1
				0x08, 0x02, // field 1: varint 2 (repeated)
				0x08, 0x03, // field 1: varint 3 (repeated)
			},
		},
		{
			name: "deeply_nested",
			data: []byte{
				0x1a, 0x07, // field 3: length-delimited (7 bytes)
				0x1a, 0x05, // nested field 3: length-delimited (5 bytes)
				0x08, 0x96, 0x01, // deeply nested field 1: varint 150
				0x10, 0x7f, // deeply nested field 2: varint 127
			},
		},
		{
			name: "single_field_message",
			data: []byte{0x08, 0x82, 0xbe, 0x02}, // field 1: varint 40706 (like in msg-b)
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := DecodeRaw(tc.data)
			require.NoError(t, err, "DecodeRaw shouldn't have failed for %s", tc.name)

			cmd := exec.Command("protoc", "--decode_raw")
			cmd.Stdin = bytes.NewReader(tc.data)
			protocOutput, err := cmd.Output()
			require.NoError(t, err, "protoc --decode_raw shouldn't have failed for %s", tc.name)

			assert.Equal(t, protocOutput, got, "output should've matched protoc's output for %s", tc.name)
		})
	}
}

func TestDecodeRawOnSampleMsg(t *testing.T) {
	// GIVEN
	// WHEN
	got, err := DecodeRaw(sampleMsg)

	// THEN
	require.NoError(t, err)
	snaps.MatchStandaloneSnapshot(t, string(got))
}

func TestDecodeRawFailsOnIncorrectInput(t *testing.T) {
	// GIVEN
	// WHEN
	data := []byte("this is not a protobuf encoded message")
	_, err := DecodeRaw(data)

	// THEN
	assert.ErrorIs(t, err, errWireDataIsMalformed)
}
