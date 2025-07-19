# Go Tools Package

This repository is used to host shared Go packages that are commonly used across multiple  services.

## 📦 Packages

### 📝 logtool

A logging utility that integrates [Uber's Zap](https://github.com/uber-go/zap) with the [Fiber](https://gofiber.io/) web framework. It provides:

- JSON logs for production
- Pretty logs for development
- Sentry integration using structured log fields as tags

The implementation has been merged with the `main` branch of `soren-notification` and is currently deployed on our server.

We use [GlitchTip](https://glitchtip.com/documentation/install) (an open-source Sentry-compatible tool) hosted on our own infrastructure for capturing and managing errors.

---

### How to use

Import `logtool` in your project:

```go
import "github.com/mohajel/go-tools-package/logtool"
```

Initialize it in your application startup:

```go
const (
	ServiceName = "your-service-name"
)

func init() {
	if config.GetDeveloperModeLogEnabled() {
		logtool.Init(ServiceName, true)
	} else {
		logtool.InitWithSentry(ServiceName, config.GetSentryDSN())
	}
	logtool.GetLogger().Warnf("Starting %s on port %s", ServiceName, config.GetPort())
}
```
Then, you can use the logger in your application.

Logs will be output in JSON format in production and pretty format in development mode.

You can log messages at different levels.

Logs that are warns and errors will be sent to Sentry for error tracking.

```go
logtool.GetLogger().Info("This is an info message")
logtool.GetLogger().Error("This is an error message")
```
You can also use the logger to log structured data:

```go
logtool.GetLogger().Warnw("This is a warning message", "key1", "value1", "key2", "value2")
```

Use the Fiber middleware to log HTTP requests:
```go
import "github.com/mohajel/go-tools-package/logtool"
app.Use(logtool.FiberZapLogger())
```

When API req has error and returns status code bigger than 400, the error will be logged with the request details with warn level and the error message will be sent to Sentry.