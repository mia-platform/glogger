package fake

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFakeLogger(t *testing.T) {
	t.Run("no fields", func(t *testing.T) {
		t.Run("info log", func(t *testing.T) {
			logger := GetLogger()

			logger.Info("my msg")

			records := logger.OriginalLogger().AllRecords()
			require.Len(t, records, 1)
			require.Equal(t, []Record{
				{
					Level:   "info",
					Message: "my msg",
					Fields:  map[string]any{},
				},
			}, records)
		})

		t.Run("trace log", func(t *testing.T) {
			logger := GetLogger()

			logger.Trace("my msg")

			records := logger.OriginalLogger().AllRecords()
			require.Len(t, records, 1)
			require.Equal(t, []Record{
				{
					Level:   "trace",
					Message: "my msg",
					Fields:  map[string]any{},
				},
			}, records)
		})

		t.Run("more logs", func(t *testing.T) {
			logger := GetLogger()

			logger.Info("my msg")
			logger.Trace("some other")
			logger.Info("yeah")

			records := logger.OriginalLogger().AllRecords()
			require.Len(t, records, 3)
			require.Equal(t, []Record{
				{
					Level:   "info",
					Message: "my msg",
					Fields:  map[string]any{},
				},
				{
					Level:   "trace",
					Message: "some other",
					Fields:  map[string]any{},
				},
				{
					Level:   "info",
					Message: "yeah",
					Fields:  map[string]any{},
				},
			}, records)
		})
	})

	t.Run("with fields", func(t *testing.T) {
		expectedFields := map[string]any{
			"k1": "v1",
			"k2": "v2",
		}

		t.Run("info log", func(t *testing.T) {
			logger := GetLogger()

			logger.WithFields(expectedFields).Info("my msg")

			records := logger.OriginalLogger().AllRecords()
			require.Len(t, records, 1)
			require.Equal(t, []Record{
				{
					Level:   "info",
					Message: "my msg",
					Fields:  expectedFields,
				},
			}, records)
		})

		t.Run("trace log", func(t *testing.T) {
			logger := GetLogger()

			logger.WithFields(expectedFields).Trace("my msg")

			records := logger.OriginalLogger().AllRecords()
			require.Len(t, records, 1)
			require.Equal(t, []Record{
				{
					Level:   "trace",
					Message: "my msg",
					Fields:  expectedFields,
				},
			}, records)
		})

		t.Run("more logs", func(t *testing.T) {
			logger := GetLogger()

			logger.WithFields(expectedFields).Info("my msg")
			logger.WithFields(map[string]any{
				"some": "value",
			}).Trace("some other")
			logger.WithFields(expectedFields).Info("yeah")

			records := logger.OriginalLogger().AllRecords()
			require.Len(t, records, 3)
			require.Equal(t, []Record{
				{
					Level:   "info",
					Message: "my msg",
					Fields:  expectedFields,
				},
				{
					Level:   "trace",
					Message: "some other",
					Fields: map[string]any{
						"some": "value",
					},
				},
				{
					Level:   "info",
					Message: "yeah",
					Fields:  expectedFields,
				},
			}, records)
		})

		t.Run("more logs with separate loggers", func(t *testing.T) {
			logger := GetLogger()

			l1 := logger.WithFields(expectedFields)
			l1.Info("my msg")
			l1.WithFields(map[string]any{
				"some": "value",
			}).Trace("some other")

			logger.WithFields(map[string]any{
				"a": "b",
			}).Info("yeah")

			records := logger.OriginalLogger().AllRecords()
			require.Len(t, records, 3)
			require.Equal(t, []Record{
				{
					Level:   "info",
					Message: "my msg",
					Fields:  expectedFields,
				},
				{
					Level:   "trace",
					Message: "some other",
					Fields: map[string]any{
						"k1":   "v1",
						"k2":   "v2",
						"some": "value",
					},
				},
				{
					Level:   "info",
					Message: "yeah",
					Fields: map[string]any{
						"a": "b",
					},
				},
			}, records)
		})
	})
}
