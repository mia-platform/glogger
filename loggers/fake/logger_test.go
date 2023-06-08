package fake

import (
	"context"
	"testing"

	"github.com/mia-platform/glogger/v3"
	"github.com/stretchr/testify/require"
)

func TestFakeLogger(t *testing.T) {
	t.Run("no fields", func(t *testing.T) {
		t.Run("info log", func(t *testing.T) {
			logger := GetLogger()

			logger.Info("my msg")

			entries := logger.GetOriginalLogger()
			require.Len(t, entries, 1)
			require.Equal(t, []Entry{
				{
					Level:   "info",
					Message: "my msg",
					Fields:  map[string]any{},
				},
			}, entries)
		})

		t.Run("trace log", func(t *testing.T) {
			logger := GetLogger()

			logger.Trace("my msg")

			entries := logger.GetOriginalLogger()
			require.Len(t, entries, 1)
			require.Equal(t, []Entry{
				{
					Level:   "trace",
					Message: "my msg",
					Fields:  map[string]any{},
				},
			}, entries)
		})

		t.Run("more logs", func(t *testing.T) {
			logger := GetLogger()

			logger.Info("my msg")
			logger.Trace("some other")
			logger.Info("yeah")

			entries := logger.GetOriginalLogger()
			require.Len(t, entries, 3)
			require.Equal(t, []Entry{
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
			}, entries)
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

			entries := logger.GetOriginalLogger()
			require.Len(t, entries, 1)
			require.Equal(t, []Entry{
				{
					Level:   "info",
					Message: "my msg",
					Fields:  expectedFields,
				},
			}, entries)
		})

		t.Run("trace log", func(t *testing.T) {
			logger := GetLogger()

			logger.WithFields(expectedFields).Trace("my msg")

			entries := logger.GetOriginalLogger()
			require.Len(t, entries, 1)
			require.Equal(t, []Entry{
				{
					Level:   "trace",
					Message: "my msg",
					Fields:  expectedFields,
				},
			}, entries)
		})

		t.Run("more logs", func(t *testing.T) {
			logger := GetLogger()

			logger.WithFields(expectedFields).Info("my msg")
			logger.WithFields(expectedFields).Trace("some other")
			logger.WithFields(expectedFields).Info("yeah")

			entries := logger.GetOriginalLogger()
			require.Len(t, entries, 3)
			require.Equal(t, []Entry{
				{
					Level:   "info",
					Message: "my msg",
					Fields:  expectedFields,
				},
				{
					Level:   "trace",
					Message: "some other",
					Fields:  expectedFields,
				},
				{
					Level:   "info",
					Message: "yeah",
					Fields:  expectedFields,
				},
			}, entries)
		})
	})

	t.Run("get from context", func(t *testing.T) {
		ctx := context.Background()
		logger := GetLogger()

		ctx = glogger.WithLogger(ctx, logger)

		require.NotNil(t, GetFromContext(ctx))
	})

	t.Run("get from context panic if not found", func(t *testing.T) {
		ctx := context.Background()

		require.PanicsWithError(t, "logger not in context", func() { GetFromContext(ctx) })
	})
}
