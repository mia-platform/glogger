# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased

### Fixed
- [BMP-507](https://makeitapp.atlassian.net/browse/BMP-507): using `x-forwarded-for` header to get IP in logs

### Changes
- [BMP-507](https://makeitapp.atlassian.net/browse/BMP-507): now host field is compliant with Mia-Platform logs
- [BMP-507](https://makeitapp.atlassian.net/browse/BMP-507): now responseTime has been converted from second to millisecond

## v2.0.0 - 29/09/2020

### BREAKING CHANGES

- Request and response logged information are now compliant with Mia-Platform logging guidelines. To see the guidelines, please check [Mia Platform Docs](https://docs.mia-platform.eu/docs/development_suite/monitoring-dashboard/dev_ops_guide/log). You can find the implementation details [here](https://github.com/mia-platform/glogger/blob/master/logmiddleware.go).

## 1.0.0 - 10/12/2019

- Initial Release 🎉🎉🎉
