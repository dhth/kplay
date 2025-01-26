package serde

import (
	"errors"
	"fmt"

	"github.com/tidwall/pretty"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

var (
	errCouldntUnmarshalProtoMsg     = errors.New("couldn't unmarshal protobuf encoded message")
	errCouldntConvertProtoMsgToJSON = errors.New("couldn't convert proto message to JSON")
)

func ParseProtobufEncodedBytes(bytes []byte, msgDescriptor protoreflect.MessageDescriptor) ([]byte, error) {
	msg := dynamicpb.NewMessage(msgDescriptor)

	if err := proto.Unmarshal(bytes, msg); err != nil {
		return nil, fmt.Errorf("%w: %s", errCouldntUnmarshalProtoMsg, err.Error())
	}

	jsonBytes, err := protojson.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errCouldntConvertProtoMsgToJSON, err.Error())
	}

	return pretty.Pretty(jsonBytes), nil
}
