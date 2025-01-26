package kafka

import (
	"errors"
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

var (
	errCouldntUnmarshallDescriptorSet  = errors.New("couldn't unmarshall descriptor set file contents")
	errCouldntCreateProtoRegistryFiles = errors.New("couldn't create proto registry files from descriptor set")
	errDescriptorNameIsInvalid         = errors.New("descriptor name is invalid")
	errCouldntFindDescriptor           = errors.New("couldn't find descriptor by name")
)

func GetDescriptorFromDescriptorSet(descSetBytes []byte, descName string) (protoreflect.MessageDescriptor, error) {
	descriptorName := protoreflect.FullName(descName)
	if !descriptorName.IsValid() {
		return nil, errDescriptorNameIsInvalid
	}

	var fds descriptorpb.FileDescriptorSet
	err := proto.Unmarshal(descSetBytes, &fds)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errCouldntUnmarshallDescriptorSet, err.Error())
	}

	files, err := protodesc.NewFiles(&fds)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errCouldntCreateProtoRegistryFiles, err.Error())
	}

	reg := protoregistry.GlobalFiles
	files.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		err := reg.RegisterFile(fd)
		return err == nil
	})

	descriptor, err := reg.FindDescriptorByName(descriptorName)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errCouldntFindDescriptor, err.Error())
	}

	return descriptor.(protoreflect.MessageDescriptor), nil
}
