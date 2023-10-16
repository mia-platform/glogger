package logrus

import (
	"context"
	"testing"

	"github.com/mia-platform/glogger/v4"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestLogger(t *testing.T) {
	t.Run("no fields", func(t *testing.T) {
		t.Run("info log", func(t *testing.T) {
			logrusLogger, hook := test.NewNullLogger()

			logger := GetLogger(logrus.NewEntry(logrusLogger))

			logger.Info("my msg")

			require.Len(t, hook.AllEntries(), 1)
			assertLog(t, hook.LastEntry(), expectedLog{
				Level:   "info",
				Message: "my msg",
				Fields:  map[string]any{},
			})
		})

		t.Run("trace log", func(t *testing.T) {
			logrusLogger, hook := test.NewNullLogger()
			logrusLogger.SetLevel(logrus.TraceLevel)

			logger := GetLogger(logrus.NewEntry(logrusLogger))

			logger.Trace("my msg")

			require.Len(t, hook.AllEntries(), 1)
			assertLog(t, hook.LastEntry(), expectedLog{
				Level:   "trace",
				Message: "my msg",
				Fields:  map[string]any{},
			})
		})

		t.Run("more logs", func(t *testing.T) {
			logrusLogger, hook := test.NewNullLogger()
			logrusLogger.SetLevel(logrus.TraceLevel)

			logger := GetLogger(logrus.NewEntry(logrusLogger))

			logger.Info("my msg")
			logger.Trace("some other")
			logger.Info("yeah")

			require.Len(t, hook.AllEntries(), 3)
			assertLog(t, hook.AllEntries()[0], expectedLog{
				Level:   "info",
				Message: "my msg",
				Fields:  map[string]any{},
			})
			assertLog(t, hook.AllEntries()[1], expectedLog{
				Level:   "trace",
				Message: "some other",
				Fields:  map[string]any{},
			})
			assertLog(t, hook.LastEntry(), expectedLog{
				Level:   "info",
				Message: "yeah",
				Fields:  map[string]any{},
			})
		})
	})

	t.Run("with fields", func(t *testing.T) {
		expectedFields := map[string]any{
			"k1": "v1",
			"k2": "v2",
		}

		t.Run("info log", func(t *testing.T) {
			logrusLogger, hook := test.NewNullLogger()

			logger := GetLogger(logrus.NewEntry(logrusLogger))

			logger.WithFields(expectedFields).Info("my msg")

			require.Len(t, hook.AllEntries(), 1)
			assertLog(t, hook.LastEntry(), expectedLog{
				Level:   "info",
				Message: "my msg",
				Fields:  expectedFields,
			})
		})

		t.Run("trace log", func(t *testing.T) {
			logrusLogger, hook := test.NewNullLogger()
			logrusLogger.SetLevel(logrus.TraceLevel)

			logger := GetLogger(logrus.NewEntry(logrusLogger))

			logger.WithFields(expectedFields).Trace("my msg")

			require.Len(t, hook.AllEntries(), 1)
			assertLog(t, hook.LastEntry(), expectedLog{
				Level:   "trace",
				Message: "my msg",
				Fields:  expectedFields,
			})
		})

		t.Run("more logs", func(t *testing.T) {
			logrusLogger, hook := test.NewNullLogger()
			logrusLogger.SetLevel(logrus.TraceLevel)

			logger := GetLogger(logrus.NewEntry(logrusLogger))

			logger.WithFields(expectedFields).Info("my msg")
			logger.WithFields(expectedFields).Trace("some other")
			logger.WithFields(expectedFields).Info("yeah")
			logger.Info("ok")

			require.Len(t, hook.AllEntries(), 4)
			assertLog(t, hook.AllEntries()[0], expectedLog{
				Level:   "info",
				Message: "my msg",
				Fields:  expectedFields,
			})
			assertLog(t, hook.AllEntries()[1], expectedLog{
				Level:   "trace",
				Message: "some other",
				Fields:  expectedFields,
			})
			assertLog(t, hook.AllEntries()[2], expectedLog{
				Level:   "info",
				Message: "yeah",
				Fields:  expectedFields,
			})
			assertLog(t, hook.LastEntry(), expectedLog{
				Level:   "info",
				Message: "ok",
				Fields:  map[string]any{},
			})
		})
	})

	t.Run("with context", func(t *testing.T) {
		ctx := context.Background()
		type ctxKey struct{}
		ctx = context.WithValue(ctx, ctxKey{}, "foo")

		expectedFields := map[string]any{
			"k1": "v1",
			"k2": "v2",
		}

		t.Run("info log", func(t *testing.T) {
			logrusLogger, hook := test.NewNullLogger()

			logger := GetLogger(logrus.NewEntry(logrusLogger))

			logger.WithContext(ctx).Info("my msg")

			require.Len(t, hook.AllEntries(), 1)
			assertLog(t, hook.LastEntry(), expectedLog{
				Level:   "info",
				Message: "my msg",
				Context: ctx,
			})
		})

		t.Run("trace log", func(t *testing.T) {
			logrusLogger, hook := test.NewNullLogger()
			logrusLogger.SetLevel(logrus.TraceLevel)

			logger := GetLogger(logrus.NewEntry(logrusLogger))

			logger.WithContext(ctx).Trace("my msg")

			require.Len(t, hook.AllEntries(), 1)
			assertLog(t, hook.LastEntry(), expectedLog{
				Level:   "trace",
				Message: "my msg",
				Context: ctx,
			})
		})

		t.Run("context + fields", func(t *testing.T) {
			logrusLogger, hook := test.NewNullLogger()
			logrusLogger.SetLevel(logrus.TraceLevel)

			logger := GetLogger(logrus.NewEntry(logrusLogger))

			logger.WithContext(ctx).WithFields(expectedFields).Info("my msg")

			require.Len(t, hook.AllEntries(), 1)
			assertLog(t, hook.LastEntry(), expectedLog{
				Level:   "info",
				Message: "my msg",
				Context: ctx,
				Fields:  expectedFields,
			})
		})

		t.Run("fields + context", func(t *testing.T) {
			logrusLogger, hook := test.NewNullLogger()
			logrusLogger.SetLevel(logrus.TraceLevel)

			logger := GetLogger(logrus.NewEntry(logrusLogger))

			logger.WithFields(expectedFields).WithContext(ctx).Info("my msg")

			require.Len(t, hook.AllEntries(), 1)
			assertLog(t, hook.LastEntry(), expectedLog{
				Level:   "info",
				Message: "my msg",
				Context: ctx,
				Fields:  expectedFields,
			})
		})

		t.Run("more logs", func(t *testing.T) {
			logrusLogger, hook := test.NewNullLogger()
			logrusLogger.SetLevel(logrus.TraceLevel)

			logger := GetLogger(logrus.NewEntry(logrusLogger))

			logger.WithFields(expectedFields).WithContext(ctx).Info("my msg")
			logger.WithFields(expectedFields).Trace("some other")
			logger.WithContext(context.Background()).WithFields(expectedFields).Info("yeah")
			logger.Info("ok")

			require.Len(t, hook.AllEntries(), 4)
			assertLog(t, hook.AllEntries()[0], expectedLog{
				Level:   "info",
				Message: "my msg",
				Fields:  expectedFields,
				Context: ctx,
			})
			assertLog(t, hook.AllEntries()[1], expectedLog{
				Level:   "trace",
				Message: "some other",
				Fields:  expectedFields,
			})
			assertLog(t, hook.AllEntries()[2], expectedLog{
				Level:   "info",
				Message: "yeah",
				Fields:  expectedFields,
				Context: context.Background(),
			})
			assertLog(t, hook.LastEntry(), expectedLog{
				Level:   "info",
				Message: "ok",
			})
		})
	})

	t.Run("save and retrieve from context", func(t *testing.T) {
		nullLogger, hook := test.NewNullLogger()
		entry := nullLogger.WithField("some", "field")

		ctx := context.Background()
		ctx = glogger.WithLogger(ctx, entry)

		actual := FromContext(ctx)
		require.NotNil(t, actual)

		actual.Info("something")
		require.Len(t, hook.AllEntries(), 1)
		require.Equal(t, hook.LastEntry().Data["some"], "field")
	})

	t.Run("get from context return default if not found in context", func(t *testing.T) {
		ctx := context.Background()

		require.NotNil(t, FromContext(ctx))
	})

	t.Run("get original logger", func(t *testing.T) {
		logrusLogger, _ := test.NewNullLogger()

		logger := GetLogger(logrus.NewEntry(logrusLogger))

		require.IsType(t, &logrus.Entry{}, logger.OriginalLogger())
	})
}

type expectedLog struct {
	Message string
	Level   string
	Fields  map[string]any
	Context context.Context
}

func assertLog(t *testing.T, logEntry *logrus.Entry, expected expectedLog) {
	t.Helper()

	if expected.Fields == nil {
		expected.Fields = map[string]any{}
	}

	require.Equal(t, expected, expectedLog{
		Level:   logEntry.Level.String(),
		Message: logEntry.Message,
		Fields:  map[string]any(logEntry.Data),
		Context: logEntry.Context,
	}, "Unexpected log data")
}
