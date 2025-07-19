package logtool

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitWithSentry(serviceName, dsn string) {
	DevMode = false
	ServiceName = serviceName
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		TracesSampleRate: 1.0,
		AttachStacktrace: true,
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}

	zapConfig := zap.NewProductionConfig()
	zapConfig.EncoderConfig.TimeKey = "time"

	encoder := zapcore.NewJSONEncoder(zapConfig.EncoderConfig)
	consoleCore := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		zapcore.InfoLevel,
	)

	sentryCore := newSentryCore(zapcore.WarnLevel)

	core := zapcore.NewTee(
		consoleCore,
		sentryCore,
	)

	logger := zap.New(core, zap.AddCaller())
	sugar = logger.Sugar()
}

type sentryCore struct {
	zapcore.LevelEnabler
}

func newSentryCore(enab zapcore.LevelEnabler) zapcore.Core {
	return &sentryCore{
		LevelEnabler: enab,
	}
}

func (c *sentryCore) With(fields []zapcore.Field) zapcore.Core {
	return c
}

func (c *sentryCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}

func (c *sentryCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	tags := make(map[string]string)
	extras := make(map[string]interface{})

	for _, field := range fields {
		switch field.Type {
		case zapcore.StringType:
			tags[field.Key] = field.String
		default:
			extras[field.Key] = field.Interface
		}
	}

	event := sentry.NewEvent()
	event.Level = zapLevelToSentryLevel(ent.Level)

	uniqueID := uuid.New().String()
	event.Message = fmt.Sprintf("[%s] - %s - %s", uniqueID, ServiceName, ent.Message)
	event.Timestamp = ent.Time

	extras["logger"] = "zap"
	extras["level"] = ent.Level.String()
	extras["time"] = ent.Time.Format(time.RFC3339)
	extras["pid"] = os.Getpid()

	if ent.Caller.Defined {
		callerStr := fmt.Sprintf("%s:%d", ent.Caller.File, ent.Caller.Line)
		extras["caller"] = callerStr
	}

	if len(tags) > 0 {
		event.Tags = tags
	}
	if len(extras) > 0 {
		event.Extra = extras
	}

	if ent.Level >= zapcore.ErrorLevel {
		event.Threads = []sentry.Thread{{
			Stacktrace: sentry.NewStacktrace(),
			Crashed:    false,
			Current:    true,
		}}
	}

	sentry.CaptureEvent(event)
	return nil
}

func (c *sentryCore) Sync() error {
	sentry.Flush(2 * time.Second)
	return nil
}

func zapLevelToSentryLevel(level zapcore.Level) sentry.Level {
	switch level {
	case zapcore.DebugLevel:
		return sentry.LevelDebug
	case zapcore.InfoLevel:
		return sentry.LevelInfo
	case zapcore.WarnLevel:
		return sentry.LevelWarning
	case zapcore.ErrorLevel, zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		return sentry.LevelError
	default:
		return sentry.LevelInfo
	}
}
