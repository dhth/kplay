package cmd

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	errInvalidPartitionOffsetFormat = errors.New("invalid partition:offset format provided")
	errInvalidPartitionProvided     = errors.New("invalid partition number provided")
	errInvalidOffsetProvided        = errors.New("invalid offset provided")
)

func parseFromOffset(value string) (*int64, map[int32]int64, error) {
	if strings.Contains(value, ":") {
		partitionOffsets := make(map[int32]int64)
		pairs := strings.SplitSeq(value, ",")
		for pair := range pairs {
			parts := strings.Split(strings.TrimSpace(pair), ":")
			if len(parts) != 2 {
				return nil, nil, fmt.Errorf("%w: %q", errInvalidPartitionOffsetFormat, pair)
			}

			partition, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 32)
			if err != nil {
				return nil, nil, fmt.Errorf("%w (%q): %w", errInvalidPartitionProvided, pair, err)
			}

			offset, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
			if err != nil {
				return nil, nil, fmt.Errorf("%w (%q): %w", errInvalidOffsetProvided, pair, err)
			}

			partitionOffsets[int32(partition)] = offset
		}

		return nil, partitionOffsets, nil
	}

	offset, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil, nil, fmt.Errorf("%w (%q): %w", errInvalidOffsetProvided, value, err)
	}

	return &offset, nil, nil
}
