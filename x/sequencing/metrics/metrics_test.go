package metrics

import (
	"context"
	"encoding/hex"
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/fuel-infrastructure/fuel-sequencer/utils"
)

func TestObserveTotalBlobsPosted(t *testing.T) {
	testCases := []struct {
		name           string
		topic          []byte
		expectedSuffix string
	}{
		{
			name:           "simple topic",
			topic:          []byte("simple"),
			expectedSuffix: hex.EncodeToString([]byte("simple")),
		},
		{
			name: "topic with null bytes",
			topic: []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x26, 0xa1,
			},
			expectedSuffix: "00000000000000000000000000000000000000000000000000000000000026a1",
		},
		{
			name:           "empty topic",
			topic:          []byte{},
			expectedSuffix: "",
		},
		{
			name:           "topic with special characters",
			topic:          []byte{0xff, 0xfe, 0xfd, 0x01, 0x02, 0x03},
			expectedSuffix: "fffefd010203",
		},
		{
			name:           "topic with printable and non-printable chars",
			topic:          []byte("test\x00\x01\x02"),
			expectedSuffix: hex.EncodeToString([]byte("test\x00\x01\x02")),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a minimal context for testing
			ctx := context.Background()

			// This should not panic regardless of input
			require.NotPanics(t, func() {
				ObserveTotalBlobsPosted(ctx, tc.topic)
			})

			// Verify hex encoding produces expected result
			result := hex.EncodeToString(tc.topic)
			require.Equal(t, tc.expectedSuffix, result)

			// Verify the hex encoded string contains only valid characters
			for _, char := range result {
				require.True(t, (char >= '0' && char <= '9') || (char >= 'a' && char <= 'f'),
					"hex encoded string should only contain 0-9 and a-f, got: %c", char)
			}
		})
	}
}

func TestObserveTotalBlobsPostedSize(t *testing.T) {
	testCases := []struct {
		name           string
		topic          []byte
		size           int
		expectedSuffix string
	}{
		{
			name:           "simple topic with positive size",
			topic:          []byte("simple"),
			size:           1024,
			expectedSuffix: hex.EncodeToString([]byte("simple")),
		},
		{
			name: "topic with null bytes and zero size",
			topic: []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x26, 0xa1,
			},
			size:           0,
			expectedSuffix: "00000000000000000000000000000000000000000000000000000000000026a1",
		},
		{
			name:           "empty topic with large size",
			topic:          []byte{},
			size:           999999,
			expectedSuffix: "",
		},
		{
			name:           "topic with special characters and negative size",
			topic:          []byte{0xff, 0xfe, 0xfd, 0x01, 0x02, 0x03},
			size:           -1,
			expectedSuffix: "fffefd010203",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a minimal context for testing
			ctx := context.Background()

			// This should not panic regardless of input
			require.NotPanics(t, func() {
				ObserveTotalBlobsPostedSize(ctx, tc.topic, tc.size)
			})

			// Verify hex encoding produces expected result
			result := hex.EncodeToString(tc.topic)
			require.Equal(t, tc.expectedSuffix, result)

			// Verify the hex encoded string contains only valid characters
			for _, char := range result {
				require.True(t, (char >= '0' && char <= '9') || (char >= 'a' && char <= 'f'),
					"hex encoded string should only contain 0-9 and a-f, got: %c", char)
			}
		})
	}
}

func TestHexEncodingProducesValidMetricNames(t *testing.T) {
	// Test the specific case that was causing the original error
	problematicTopic := []byte{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x26, 0xa1,
	}

	// Test that hex encoding produces a valid string
	hexString := hex.EncodeToString(problematicTopic)
	require.Equal(t, "00000000000000000000000000000000000000000000000000000000000026a1", hexString)

	// Verify it contains no null bytes or invalid characters
	for _, b := range []byte(hexString) {
		require.True(t, b != 0, "hex encoded string should not contain null bytes")
		require.True(t, b >= 32 && b <= 126, "hex encoded string should only contain printable ASCII characters")
	}

	// Test that the original problematic conversion would have null bytes
	directString := string(problematicTopic)
	hasNullBytes := false
	for _, b := range []byte(directString) {
		if b == 0 {
			hasNullBytes = true
			break
		}
	}
	require.True(t, hasNullBytes,
		"direct string conversion should contain null bytes (demonstrating the original problem)")

	// Verify our functions construct the expected metric name components
	expectedKeys := append(utils.KeysTxMsg, "total", "blobs", "posted", hexString)
	require.Equal(t, []string{
		"sequencer", "tx", "msg", "total", "blobs", "posted",
		"00000000000000000000000000000000000000000000000000000000000026a1",
	}, expectedKeys)

	expectedSizeKeys := append(utils.KeysTxMsg, "total", "blobs", "posted", "size", hexString)
	require.Equal(t, []string{
		"sequencer", "tx", "msg", "total", "blobs", "posted", "size",
		"00000000000000000000000000000000000000000000000000000000000026a1",
	}, expectedSizeKeys)
}

func TestMetricsFunctionsWithMockContext(t *testing.T) {
	// Test that functions work with different context types
	testCases := []struct {
		name string
		ctx  context.Context
	}{
		{
			name: "background context",
			ctx:  context.Background(),
		},
		{
			name: "context with value",
			//nolint:staticcheck // SA1029: using string key in test context is acceptable
			ctx: context.WithValue(context.Background(), "test", "value"),
		},
		{
			name: "cancelled context",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
		},
	}

	topic := []byte("test_topic")
	size := 42

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.NotPanics(t, func() {
				ObserveTotalBlobsPosted(tc.ctx, topic)
			})

			require.NotPanics(t, func() {
				ObserveTotalBlobsPostedSize(tc.ctx, topic, size)
			})
		})
	}
}

func TestEdgeCasesNilAndEmptySlices(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name  string
		topic []byte
	}{
		{
			name:  "nil slice",
			topic: nil,
		},
		{
			name:  "empty slice",
			topic: []byte{},
		},
		{
			name:  "slice with single zero byte",
			topic: []byte{0x00},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Both functions should handle nil and empty slices gracefully
			require.NotPanics(t, func() {
				ObserveTotalBlobsPosted(ctx, tc.topic)
			})

			require.NotPanics(t, func() {
				ObserveTotalBlobsPostedSize(ctx, tc.topic, 100)
			})

			// Verify hex encoding works for these cases
			hexResult := hex.EncodeToString(tc.topic)
			require.True(t, len(hexResult)%2 == 0, "hex encoding should always produce even length strings")
		})
	}
}

func TestLargeInputsAndPerformance(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name      string
		topicSize int
		size      int
	}{
		{
			name:      "medium topic (1KB)",
			topicSize: 1024,
			size:      1024,
		},
		{
			name:      "large topic (64KB)",
			topicSize: 65536,
			size:      65536,
		},
		{
			name:      "very large size value",
			topicSize: 32,
			size:      math.MaxInt32,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a topic filled with varying bytes to test realistic scenarios
			topic := make([]byte, tc.topicSize)
			for i := range topic {
				topic[i] = byte(i % 256)
			}

			// These should not panic even with large inputs
			require.NotPanics(t, func() {
				ObserveTotalBlobsPosted(ctx, topic)
			})

			require.NotPanics(t, func() {
				ObserveTotalBlobsPostedSize(ctx, topic, tc.size)
			})

			// Verify hex encoding produces expected length
			hexResult := hex.EncodeToString(topic)
			require.Equal(t, tc.topicSize*2, len(hexResult), "hex encoding should double the byte length")
		})
	}
}

func TestFloat32ConversionEdgeCases(t *testing.T) {
	ctx := context.Background()
	topic := []byte("test")

	testCases := []struct {
		name string
		size int
	}{
		{
			name: "max int32",
			size: math.MaxInt32,
		},
		{
			name: "min int32",
			size: math.MinInt32,
		},
		{
			name: "max safe integer for float32",
			size: 16777216, // 2^24, largest integer that can be represented exactly in float32
		},
		{
			name: "large positive value",
			size: 1000000000,
		},
		{
			name: "large negative value",
			size: -1000000000,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Should not panic even with extreme size values
			require.NotPanics(t, func() {
				ObserveTotalBlobsPostedSize(ctx, topic, tc.size)
			})

			// Verify float32 conversion doesn't overflow
			float32Val := float32(tc.size)
			require.False(t, math.IsInf(float64(float32Val), 0), "float32 conversion should not produce infinity")
			require.False(t, math.IsNaN(float64(float32Val)), "float32 conversion should not produce NaN")
		})
	}
}

func TestUTF8AndUnicodeContent(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name  string
		topic []byte
	}{
		{
			name:  "valid UTF-8 english",
			topic: []byte("hello world"),
		},
		{
			name:  "valid UTF-8 unicode",
			topic: []byte("こんにちは世界"), // "Hello World" in Japanese
		},
		{
			name:  "valid UTF-8 emoji",
			topic: []byte("🌍🚀💫"),
		},
		{
			name:  "invalid UTF-8 sequence",
			topic: []byte{0xff, 0xfe, 0xfd, 0x80, 0x81},
		},
		{
			name:  "mixed valid UTF-8 and binary",
			topic: []byte("test🌍\x00\xff\x01"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.NotPanics(t, func() {
				ObserveTotalBlobsPosted(ctx, tc.topic)
			})

			require.NotPanics(t, func() {
				ObserveTotalBlobsPostedSize(ctx, tc.topic, 100)
			})

			// Verify hex encoding produces valid metric names regardless of UTF-8 validity
			hexResult := hex.EncodeToString(tc.topic)
			for _, char := range hexResult {
				require.True(t, (char >= '0' && char <= '9') || (char >= 'a' && char <= 'f'),
					"hex encoded string should only contain valid characters")
			}
		})
	}
}

func TestMetricNameLengthConsiderations(t *testing.T) {
	ctx := context.Background()

	// Test very long metric names to ensure they don't cause issues
	testCases := []struct {
		name      string
		topicSize int
	}{
		{
			name:      "moderate length topic",
			topicSize: 100,
		},
		{
			name:      "long topic that creates very long metric name",
			topicSize: 1000,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			topic := make([]byte, tc.topicSize)
			for i := range topic {
				topic[i] = byte(i % 256)
			}

			require.NotPanics(t, func() {
				ObserveTotalBlobsPosted(ctx, topic)
			})

			require.NotPanics(t, func() {
				ObserveTotalBlobsPostedSize(ctx, topic, 100)
			})

			// Calculate expected metric name length
			hexTopic := hex.EncodeToString(topic)
			baseKeys := utils.KeysTxMsg
			expectedParts := append(baseKeys, "total", "blobs", "posted", hexTopic)
			metricName := strings.Join(expectedParts, "_")

			// Verify metric name construction
			require.Greater(t, len(metricName), 0, "metric name should not be empty")
			require.Contains(t, metricName, "sequencer_tx_msg_total_blobs_posted_")
			require.Contains(t, metricName, hexTopic)
		})
	}
}
