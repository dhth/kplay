package serde

import (
	"errors"
	"fmt"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

const (
	maxRecursionDepth = 100
	maxFieldNumber    = 536870911 // protobuf maximum field number (2^29 - 1)
)

var (
	errCouldntUnmarshalProtoMsg     = errors.New("couldn't unmarshal protobuf encoded message")
	errCouldntConvertProtoMsgToJSON = errors.New("couldn't convert proto message to JSON")
	errWireDataIsMalformed          = errors.New("wire data is malformed")
)

type rawDecoder struct {
	result      *strings.Builder
	indentLevel int
}

func TranscodeProto(bytes []byte, msgDescriptor protoreflect.MessageDescriptor) ([]byte, error) {
	msg := dynamicpb.NewMessage(msgDescriptor)

	err := proto.Unmarshal(bytes, msg)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errCouldntUnmarshalProtoMsg, err.Error())
	}

	marshallOptions := protojson.MarshalOptions{
		Indent: "  ",
	}
	jsonBytes, err := marshallOptions.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errCouldntConvertProtoMsgToJSON, err.Error())
	}

	return jsonBytes, nil
}

func DecodeRaw(data []byte) ([]byte, error) {
	decoder := newRawDecoder()
	err := decoder.writeRawTagValuePairs(data, maxRecursionDepth)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errWireDataIsMalformed, err.Error())
	}

	return []byte(decoder.string()), nil
}

func (d *rawDecoder) writeRawTagValuePairs(data []byte, recursionBudget int) error {
	remaining := data

	for len(remaining) > 0 {
		fieldNum, wireType, n := protowire.ConsumeTag(remaining)
		if n < 0 {
			return fmt.Errorf("failed to consume tag: %w", protowire.ParseError(n))
		}
		remaining = remaining[n:]

		indent := d.getIndent()

		switch wireType {
		case protowire.VarintType:
			value, n := protowire.ConsumeVarint(remaining)
			if n < 0 {
				return fmt.Errorf("failed to consume varint: %w", protowire.ParseError(n))
			}
			d.printLiteral(fmt.Sprintf("%s%d", indent, fieldNum))
			d.printLiteral(": ")
			d.printLiteral(fmt.Sprintf("%d", value))
			d.printLineEnding()
			remaining = remaining[n:]

		case protowire.Fixed32Type:
			value, n := protowire.ConsumeFixed32(remaining)
			if n < 0 {
				return fmt.Errorf("failed to consume fixed32: %w", protowire.ParseError(n))
			}
			d.printLiteral(fmt.Sprintf("%s%d", indent, fieldNum))
			d.printLiteral(": 0x")
			d.printLiteral(fmt.Sprintf("%08x", value))
			d.printLineEnding()
			remaining = remaining[n:]

		case protowire.Fixed64Type:
			value, n := protowire.ConsumeFixed64(remaining)
			if n < 0 {
				return fmt.Errorf("failed to consume fixed64: %w", protowire.ParseError(n))
			}
			d.printLiteral(fmt.Sprintf("%s%d", indent, fieldNum))
			d.printLiteral(": 0x")
			d.printLiteral(fmt.Sprintf("%016x", value))
			d.printLineEnding()
			remaining = remaining[n:]

		case protowire.BytesType:
			value, n := protowire.ConsumeBytes(remaining)
			if n < 0 {
				return fmt.Errorf("failed to consume bytes: %w", protowire.ParseError(n))
			}

			d.printLiteral(fmt.Sprintf("%s%d", indent, fieldNum))

			if len(value) > 0 && recursionBudget > 0 && canParseAsMessage(value) {
				d.printLiteral(" {\n")
				d.indent()
				err := d.writeRawTagValuePairs(value, recursionBudget-1)
				if err != nil {
					d.printLiteral(": \"")
					d.printLiteral(cEscape(value))
					d.printLiteral("\"\n")
					remaining = remaining[n:]
					continue
				}
				d.outdent()
				d.printLiteral(d.getIndent())
				d.printLiteral("}\n")
			} else {
				d.printLiteral(": \"")
				d.printLiteral(cEscape(value))
				d.printLiteral("\"\n")
			}
			remaining = remaining[n:]

		case protowire.StartGroupType:
			d.printLiteral(fmt.Sprintf("%s%d", indent, fieldNum))
			d.printLiteral(" {\n")
			d.indent()

			groupData, n := protowire.ConsumeGroup(fieldNum, remaining)
			if n < 0 {
				return fmt.Errorf("failed to consume group: %w", protowire.ParseError(n))
			}

			if recursionBudget > 0 {
				err := d.writeRawTagValuePairs(groupData, recursionBudget-1)
				if err != nil {
					d.printLiteral(": \"")
					d.printLiteral(cEscape(groupData))
					d.printLiteral("\"\n")
				}
			}

			d.outdent()
			d.printLiteral(d.getIndent())
			d.printLiteral("}\n")
			remaining = remaining[n:]

		default:
			n := protowire.ConsumeFieldValue(fieldNum, wireType, remaining)
			if n < 0 {
				return fmt.Errorf("failed to consume field value: %w", protowire.ParseError(n))
			}
			d.printLiteral(fmt.Sprintf("%s%d: <unknown wire type %d>\n", indent, fieldNum, wireType))
			remaining = remaining[n:]
		}
	}

	return nil
}

func cEscape(data []byte) string {
	var result strings.Builder

	for _, b := range data {
		switch b {
		case '\\': // backslash
			result.WriteString("\\\\")
		case '"': // double quote
			result.WriteString("\\\"")
		case '\n': // newline
			result.WriteString("\\n")
		case '\r': // carriage return
			result.WriteString("\\r")
		case '\t': // tab
			result.WriteString("\\t")
		case '\a': // bell
			result.WriteString("\\a")
		case '\b': // backspace
			result.WriteString("\\b")
		case '\f': // form feed
			result.WriteString("\\f")
		case '\v': // vertical tab
			result.WriteString("\\v")
		default:
			if b >= 32 && b <= 126 {
				// Printable ASCII character
				result.WriteByte(b)
			} else {
				// Non-printable character - use octal escape
				result.WriteString(fmt.Sprintf("\\%03o", b))
			}
		}
	}

	return result.String()
}

func canParseAsMessage(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	// Try to parse the entire message by consuming all fields
	remaining := data
	for len(remaining) > 0 {
		// Try to consume a complete field
		fieldNum, wireType, n := protowire.ConsumeTag(remaining)
		if n < 0 {
			return false
		}

		// Validate field number is in valid range
		if fieldNum == 0 || fieldNum > maxFieldNumber {
			return false
		}

		remaining = remaining[n:]

		// Try to consume the field value
		n = protowire.ConsumeFieldValue(fieldNum, wireType, remaining)
		if n < 0 {
			return false
		}

		remaining = remaining[n:]
	}

	// If we successfully consumed all data, it's a valid message
	return true
}

func newRawDecoder() *rawDecoder {
	return &rawDecoder{
		result: &strings.Builder{},
	}
}

func (d *rawDecoder) string() string {
	return d.result.String()
}

func (d *rawDecoder) indent() {
	d.indentLevel++
}

func (d *rawDecoder) outdent() {
	if d.indentLevel > 0 {
		d.indentLevel--
	}
}

func (d *rawDecoder) printLiteral(text string) {
	d.result.WriteString(text)
}

func (d *rawDecoder) getIndent() string {
	return strings.Repeat("  ", d.indentLevel)
}

func (d *rawDecoder) printLineEnding() {
	d.result.WriteString("\n")
}
