package logtool

import (
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	ServiceName string
	DevMode     bool
	sugar       *zap.SugaredLogger
)

func GetLogger() *zap.SugaredLogger {
	return sugar
}

func Init(serviceName string, devMode bool) {
	ServiceName = serviceName
	DevMode = devMode
	var logger *zap.Logger
	var err error
	if devMode {
		zapConfig := zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.TimeKey = ""
		zapConfig.OutputPaths = []string{"stdout"}

		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		zapConfig.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
		zapConfig.EncoderConfig.EncodeName = zapcore.FullNameEncoder
		zapConfig.DisableStacktrace = true
		logger, err = zapConfig.Build()
	} else {
		zapConfig := zap.NewProductionConfig()
		zapConfig.EncoderConfig.TimeKey = "time"
		zapConfig.Encoding = "json"
		zapConfig.OutputPaths = []string{"stdout"}
		logger, err = zapConfig.Build()
	}
	if err != nil {
		log.Fatal(err)
	}
	sugar = logger.Sugar()
}

// Custom Fiber logger middleware for zap
func FiberZapLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := c.Context().Time()
		err := c.Next()
		stop := c.Context().Time()
		status := c.Response().StatusCode()
		addBody := false
		title := http.StatusText(status)

		var logFn func(string, ...interface{})
		switch {
		case status >= 500:
			logFn = GetLogger().Errorw
			addBody = true
		case status >= 400:
			logFn = GetLogger().Warnw
			addBody = true
		default:
			logFn = GetLogger().Infow
		}

		fields := []interface{}{
			"method", c.Method(),
			"path", c.Path(),
			"status", status,
			"caller", "logtool.FiberZapLogger",
		}
		if !DevMode {
			fields = append(fields,
				"latency", stop.Sub(start).String(),
				"ip", c.IP(),
				"user_agent", c.Get("User-Agent"),
			)
		}

		if err != nil {
			fields = append(fields, "error", err.Error())
		}
		if addBody {
			fields = append(fields, "body", string(c.Response().Body()))
		}

		logFn(title, fields...)
		return err
	}
}
