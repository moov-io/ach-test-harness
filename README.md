<!--generated-from:bf1398c42b7dfe269385837d6c210cea156131bacff2e0012ec54535d76bc415 DO NOT REMOVE, DO UPDATE -->
moov-io/ach-test-harness
===

[![GoDoc](https://godoc.org/github.com/moov-io/ach-test-harness?status.svg)](https://godoc.org/github.com/moov-io/ach-test-harness)
[![Build Status](https://github.com/moov-io/ach-test-harness/workflows/Go/badge.svg)](https://github.com/moov-io/ach-test-harness/actions)
[![Coverage Status](https://codecov.io/gh/moov-io/ach-test-harness/branch/master/graph/badge.svg)](https://codecov.io/gh/moov-io/ach-test-harness)
[![Go Report Card](https://goreportcard.com/badge/github.com/moov-io/ach-test-harness)](https://goreportcard.com/report/github.com/moov-io/ach-test-harness)
[![Apache 2 licensed](https://img.shields.io/badge/license-Apache2-blue.svg)](https://raw.githubusercontent.com/moov-io/ach-test-harness/master/LICENSE)

A configurable FTP/SFTP server and Go library to interactively test ACH scenarios to replicate real world originations, returns, changes, prenotes, and transfers.

Docs: [docs](https://moov-io.github.io/ach-test-harness/) | [open api specification](api/api.yml)

## Project Status

This project is currently under development and could introduce breaking changes to reach a stable status. We are looking for community feedback so please try out our code or give us feedback!

## Getting Started

We publish a [public Docker image `moov/ach-test-harness`](https://hub.docker.com/r/moov/ach-test-harness/) from Docker Hub or use this repository. No configuration is required to serve on `:2222` and metrics at `:3333/metrics` in Prometheus format. <!-- We also have Docker images for [OpenShift](https://quay.io/repository/moov/ach-test-harness?tab=tags) published as `quay.io/moov/ach-test-harness`. -->

Pull & start the Docker image:
```
docker pull moov/ach-test-harness:latest
docker run -p 2222:2222 -p 3333:3333 moov/ach-test-harness:latest
```

Inspect your configuration file and setup some scenarios to match uploaded files.

```yaml
ACHTestHarness:
  Servers:
    FTP:
      RootPath: "./data"
      Hostname: "0.0.0.0"
      Port: 2222
      Auth:
        Username: "admin"
        Password: "secret"
      PassivePorts: "30000-30009"
      Paths:
        Files: "/outbound/"
        Return: "/returned/"
    Admin:
      Bind:
        Address: ":3333"
  Responses:
    # Entries that match both the DFIAccountNumber and TraceNumber will be returned with a R03 return code.
    - match:
        accountNumber: "12345678"
        traceNumber: "121042880000001"
      action:
        return:
          code: "R03"
```

The full config for Responses is below:

```yaml
# All populated fields must match for the action to be applied to the EntryDetail
match:
  # Match the DFIAccountNumber on the EntryDetail
  accountNumber: <string>
  amount:
    min: <integer>
    max: <integer>
    value: <integer>    # Either min AND max OR value is used
  debit: <object>       # Include this to only match on debits
  traceNumber: <string> # Exact match of TraceNumber

action:
  # Send the EntryDetail back with the following ACH change code
  correction:
    code: <string>
    data: <string>

  # Send the EntryDetail back with the following ACH return code
  return:
    code: <string>
```

## Getting Help

 channel | info
 ------- | -------
[Project Documentation](docs/README.md) | Our project documentation available online.
Twitter [@moov_io](https://twitter.com/moov_io)	| You can follow Moov.IO's Twitter feed to get updates on our project(s). You can also tweet us questions or just share blogs or stories.
[GitHub Issue](https://github.com/moov-io/ach-test-harness/issues) | If you are able to reproduce a problem please open a GitHub Issue under the specific project that caused the error.
[moov-io slack](https://slack.moov.io/) | Join our slack channel (`#ach`) to have an interactive discussion about the development of the project.

## Supported and Tested Platforms

- 64-bit Linux (Ubuntu, Debian), macOS, and Windows

## Contributing

Yes please! Please review our [Contributing guide](CONTRIBUTING.md) and [Code of Conduct](https://github.com/moov-io/ach/blob/master/CODE_OF_CONDUCT.md) to get started! Checkout our [issues for first time contributors](https://github.com/moov-io/ach-test-harness/contribute) for something to help out with.

This project uses [Go Modules](https://github.com/golang/go/wiki/Modules) and uses Go 1.14 or higher. See [Golang's install instructions](https://golang.org/doc/install) for help setting up Go. You can download the source code and we offer [tagged and released versions](https://github.com/moov-io/ach-test-harness/releases/latest) as well. We highly recommend you use a tagged release for production.

### Test Coverage

Improving test coverage is a good candidate for new contributors while also allowing the project to move more quickly by reducing regressions issues that might not be caught before a release is pushed out to our users. One great way to improve coverage is by adding edge cases and different inputs to functions (or [contributing and running fuzzers](https://github.com/dvyukov/go-fuzz)).

Tests can run processes (like sqlite databases), but should only do so locally.

## License

Apache License 2.0 See [LICENSE](LICENSE) for details.
