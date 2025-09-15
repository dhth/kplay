package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFromOffset(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		expectedOffset     *int64
		expectedPartitions map[int32]int64
		expectedError      error
	}{
		// SUCCESSES
		{
			name:               "simple offset",
			input:              "1000",
			expectedOffset:     int64Ptr(1000),
			expectedPartitions: nil,
			expectedError:      nil,
		},
		{
			name:               "single partition",
			input:              "0:1000",
			expectedOffset:     nil,
			expectedPartitions: map[int32]int64{0: 1000},
			expectedError:      nil,
		},
		{
			name:               "multiple partitions",
			input:              "0:1000,2:1500,5:2000",
			expectedOffset:     nil,
			expectedPartitions: map[int32]int64{0: 1000, 2: 1500, 5: 2000},
			expectedError:      nil,
		},
		{
			name:               "with spaces",
			input:              "0: 1000, 2: 1500",
			expectedOffset:     nil,
			expectedPartitions: map[int32]int64{0: 1000, 2: 1500},
			expectedError:      nil,
		},
		// FAILURES
		{
			name:          "empty value",
			input:         "",
			expectedError: errInvalidOffsetProvided,
		},
		{
			name:          "invalid offset",
			input:         "abc",
			expectedError: errInvalidOffsetProvided,
		},
		{
			name:          "empty partition pairs",
			input:         ":,:",
			expectedError: errInvalidPartitionProvided,
		},
		{
			name:          "empty partition",
			input:         "0:100,:1000",
			expectedError: errInvalidPartitionProvided,
		},
		{
			name:          "invalid partition format",
			input:         "0:100,1:1000:extra",
			expectedError: errInvalidPartitionOffsetFormat,
		},
		{
			name:          "invalid partition number",
			input:         "0:100,abc:1000",
			expectedError: errInvalidPartitionProvided,
		},
		{
			name:          "empty offset",
			input:         "0:100,1:",
			expectedError: errInvalidOffsetProvided,
		},
		{
			name:          "invalid partition offset",
			input:         "0:100,1:abc",
			expectedError: errInvalidOffsetProvided,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offset, partitions, err := parseFromOffset(tt.input)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				return
			}

			require.NoError(t, err)

			if tt.expectedOffset != nil {
				require.NotNil(t, offset)
				assert.Equal(t, *tt.expectedOffset, *offset)
			} else {
				assert.Nil(t, offset)
			}

			assert.Equal(t, tt.expectedPartitions, partitions)
		})
	}
}

func int64Ptr(v int64) *int64 {
	return &v
}
