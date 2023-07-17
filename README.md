<div align="center">

# Glogger

[![Build Status][github-actions-svg]][github-actions]
[![Go Report Card][go-report-card]][go-report-card-link]
[![GoDoc][godoc-svg]][godoc-link]

</div>

**Glogger is the logger for mia-platform go services.**

It uses a logging interface to integrate with loggers, and can expose middleware for all the existent routers
and use any combination of loggers and routers.

At the moment, we support:

**Loggers**:

- [logrus](https://github.com/sirupsen/logrus)

Do you want to use another logger? Please open a PR to include it in the repo!

**Routers**:

- [http gorilla mux router](https://github.com/gorilla/mux)
- [fiber](https://github.com/gofiber/fiber)

Do you want to use another router? Please open a PR to include it in the repo!

This library follow the Mia-Platform logging guidelines.

## Install

This library require golang at version >= 1.18

```sh
go get -u github.com/mia-platform/glogger/v4
```

## Example usage

### Basic logrus initialization

The allowed log level are those parsed by [logrus ParseLevel](https://godoc.org/github.com/sirupsen/logrus#ParseLevel) (e.g. panic, fatal, error, warn, warning, info, debug, trace).

```go
import glogrus "github.com/mia-platform/glogger/v4/loggers/logrus"

// Logger setup
logger, err := glogrus.InitHelper(glogrus.InitOptions{})
if err != nil {
  msg := fmt.Sprintf("An error occurred while creating the logger: %v", err)
  panic(msg)
}
```

## Middleware

### Gorilla Mux

Init log middleware for [mux router](https://github.com/gorilla/mux). This log the `incoming request` and `request completed` following the mia-platform guidelines.

```go
import (
  glogrus "github.com/mia-platform/glogger/v4/loggers/logrus"
  gmux "github.com/mia-platform/glogger/v4/middleware/mux"
)

router := mux.NewRouter()

middlewareLog := glogrus.GetLogger(logrus.NewEntry(logger))
router.Use(gmux.RequestMiddlewareLogger[*logrus.Entry](middlewareLog, []string{}))
```

and, to retrieve logger injected in request context:

```go
func(w http.ResponseWriter, req *http.Request) {
  loggerFn := glogrus.FromContext(r.Context())
  loggerFn.Info("log message")
}
```

### Fiber

With [fiber](https://github.com/gofiber/fiber), you can setup the middleware in this way:

```go
import (
  "github.com/gofiber/fiber/v2"
  glogrus "github.com/mia-platform/glogger/v4/loggers/logrus"
  gfiber "github.com/mia-platform/glogger/v4/middleware/fiber"
  "github.com/sirupsen/logrus"
)

app := fiber.New()

middlewareLog := glogrus.GetLogger(logrus.NewEntry(logger))
app.Use(gfiber.RequestMiddlewareLogger[*logrus.Entry](middlewareLog, []string{}))
```

And then retrieve it from the handler's context like this:

```go
app.Get("/", func(c *fiber.Ctx) error {
  log := glogrus.FromContext(c.Context())
  log.Info("log message")
  return nil
})
```

#### with excluded path

You can restrict the path where the logger middleware take effect using the second paramenter in middlewares. For example, this could be useful to exclude `incoming request` and `request completed` logging in path router.

Logger function is injected anyway in request context.

```go
router := mux.NewRouter()

middlewareLog := glogrus.GetLogger(logrus.NewEntry(logger))
router.Use(gmux.RequestMiddlewareLogger[*logrus.Entry](middlewareLog, []string{"/-/"}))

```

## How to log error message (example with logrus)

To log error message using default field

```go
_, err := myFn()

if err != nil {
  log := glogrus.FromContext(c.Context()).WithError(err).Error("error calling function")
}
```

## How to log custom fields (with logrus)

To log error message using default field

```go
glogrus.FromContext(c.Context()).WithField("key", "some field").Info("error calling function")

glogrus.FromContext(c.Context()).WithFields(&logrus.Fields{
  "key": "some field",
  "another-key": "something"
}).Info("log with custom fields")
```

## Contributing

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct,
and the process for submitting pull requests to us.

## Versioning

We use [SemVer][semver] for versioning. For the versions available,
see the [tags on this repository](https://github.com/mia-platform/glogger/tags).

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE.md](LICENSE.md)
file for details

[github-actions]: https://github.com/mia-platform/glogger/actions
[github-actions-svg]: https://github.com/mia-platform/glogger/workflows/Test%20and%20build/badge.svg
[godoc-svg]: https://godoc.org/github.com/mia-platform/glogger?status.svg
[godoc-link]: https://godoc.org/github.com/mia-platform/glogger
[go-report-card]: https://goreportcard.com/badge/github.com/mia-platform/glogger
[go-report-card-link]: https://goreportcard.com/report/github.com/mia-platform/glogger
[semver]: https://semver.org/
