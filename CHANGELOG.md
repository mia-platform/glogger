# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased

### Added

- add context to logger in middlewares. This could be useful when logger support hooks.

## v4.0.1 - 07-09-2023

### Fixed

- fix concurrency issue in logging

## v4.0.0 - 17-07-2023

### Added

- manage to support different loggers other than mux
- better routers middleware separation

## v3.0.1 - 20-03-2022

### Fixed

- fiber error log in request completed

## v3.0.0 - 23-12-2022

### BREAKING CHANGES

- renamed `RequestMiddlewareLogger` to `RequestGorillaMuxMiddlewareLogger` to distinguish between `fiber` and `mux`. The function has been moved to the `middleware` subpackage.

### Added

- new middleware for [fiber](https://github.com/gofiber/fiber) http framework

### Updated

- `golang@1.19`

## v2.1.3 - 11-03-2022

### Added

- Flush support

## v2.1.2 - 24-01-2022

### Fixed

- remove trace log creating request id

## v2.1.1

### Fixed

- log timestamp has precision in milliseconds instead of seconds.

## v2.1.0

### Added

- add option `DisableHTMLEscape` to parse log

## v2.0.3 - 02/10/2020

### Fixed
- using `x-forwarded-for` header to get IP in logs
- host field is compliant with Mia-Platform logs
- responseTime has been converted from second to millisecond

### Added

- forwardedHost key for `x-forwarded-host` header content

## v2.0.0 - 29/09/2020

### BREAKING CHANGES

- Request and response logged information are now compliant with Mia-Platform logging guidelines. To see the guidelines, please check [Mia Platform Docs](https://docs.mia-platform.eu/docs/development_suite/monitoring-dashboard/dev_ops_guide/log). You can find the implementation details [here](https://github.com/mia-platform/glogger/blob/master/logmiddleware.go).

## 1.0.0 - 10/12/2019

- Initial Release 🎉🎉🎉
