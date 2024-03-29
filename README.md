# Go Toolkit

[![Build Status](https://github.com/diptanw/go-toolkit/workflows/build-n-test/badge.svg)](https://github.com/diptanw/go-toolkit/actions)
[![Go Version](https://img.shields.io/static/v1?label=Go&message=1.18&color=9cf)](https://golang.org/doc/go1.18)
[![Release](https://img.shields.io/badge/Release-LATEST-brightgreen.svg)](https://github.com/diptanw/go-toolkit/releases/latest)

**Go Toolkit** is a **programming toolkit** for building services and serverless applications in Go.
The motivation has educational and training purposes. An incubator of potentially useful and lightweight libraries designed to unify the development technique and ensure consistency.

## Packages

- [auth](/auth/doc.go)
- [http](/server/doc.go)
- [logger](/logger/doc.go)
- [message](/message/doc.go)
- [retry](/retry/doc.go)
- [storage](/storage/doc.go)
- [async](/async/doc.go)
- [server](/server/doc.go)

## Versioning

API is currently unstable and there are no compatibility guarantees [(semver)](https://semver.org/). See [Go modules versioning model](https://github.com/golang/go/wiki/Modules#faqs--semantic-import-versioning).

## Contributing

Please see [CONTRIBUTING.md](/CONTRIBUTING.md).

## Testing

To run tests with code coverage, run the following command:

```sh
go test -race -coverprofile=c.out ./... && go tool cover -html=c.out
```

