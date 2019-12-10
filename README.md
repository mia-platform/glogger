<div align="center">

# Glogger

[![Build Status][github-actions-svg]][gitub-actions]
[![Go Report Card][go-report-card]][go-report-card-link]
[![GoDoc][godoc-svg]][godoc-link]

</div>

**Glogger is the logger for mia-platform go services.**

It uses [logrus](https://github.com/sirupsen/logrus) as logging library,
and implements a middleware to be used with [http gorilla mux router](https://github.com/gorilla/mux).

This library follow the Mia Platform logging guidelines.

## Install

This library require golang at version >= 1.13

```sh
go get -u github.com/mia-platform/glogger
```

## Example usage

### Basic logger initialization.

The allowed log level are those parsed by [logrus ParseLevel](https://godoc.org/github.com/sirupsen/logrus#ParseLevel) (e.g. panic, fatal, error, warn, warning, info, debug, trace).

```go
// Logger setup
log, err := glogger.InitHelper(logger.InitOptions{Level: "info"})
if err != nil {
	msg := fmt.Sprintf("An error occurred while creating the logger: %v", err)
	panic(msg)
}
```

### Setup log middleware

Init log middleware for [mux router](https://github.com/gorilla/mux). This log the `incoming request` and `request completed` following the mia-platform guidelines.

```go
r := mux.NewRouter()
r.Use(glogger.RequestMiddlewareLogger(log, nil))
```

and, to retrieve logger injected in request context:

```go
func(w http.ResponseWriter, req *http.Request) {
  loggerFn := logger.Get(req.Context())
  loggerFn.Info("log message")
}
```

#### with excluded path

You can restrict the path where the logger middleware take effect using the second paramenter in `RequestMiddlewareLogger`. For example, this could be useful to exclude `incoming request` and `request completed` logging in path router.

Logger function is injected anyway in request context.

```go
r := mux.NewRouter()
r.Use(glogger.RequestMiddlewareLogger(log, []string{"/-/"}))
```

## How to log error message

To log error message using default field

```go
_, err := myFn()

if err != nil {
  logger.Get(req.Context()).WithError(err).Error("error calling function")
}
```

## How to log custom fields

To log error message using default field

```go
logger.Get(req.Context()).WithField("key", "some field").Info("error calling function")

logger.Get(req.Context()).WithFields(&logrus.Fields{
  "key": "some field",
  "another-key": "something"
}).Info("log with custom fields")
```

## Contributing

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct,
and the process for submitting pull requests to us.

## Versioning

We use [SemVer][semver] for versioning. For the versions available,
see the [tags on this repository](https://github.com/mia-platform/terraform-google-project/tags).

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
