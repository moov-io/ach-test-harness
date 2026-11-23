<!--generated-from:bf1398c42b7dfe269385837d6c210cea156131bacff2e0012ec54535d76bc415 DO NOT REMOVE, DO UPDATE -->
[![Moov Banner Logo](https://user-images.githubusercontent.com/20115216/104214617-885b3c80-53ec-11eb-8ce0-9fc745fb5bfc.png)](https://github.com/moov-io)

<p align="center">
  <a href="https://github.com/moov-io/ach-test-harness/blob/master/docs/entry-search.md">Project Documentation</a>
  ·
  <a href="https://moov.io/blog/education/ach-test-harness-guide/">Quickstart Guide</a>
  ·
  <a href="https://slack.moov.io/">Community</a>
  ·
  <a href="https://moov.io/blog/">Blog</a>
  <br>
  <br>
</p>

[![GoDoc](https://godoc.org/github.com/moov-io/ach-test-harness?status.svg)](https://godoc.org/github.com/moov-io/ach-test-harness)
[![Build Status](https://github.com/moov-io/ach-test-harness/workflows/Go/badge.svg)](https://github.com/moov-io/ach-test-harness/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/moov-io/ach-test-harness)](https://goreportcard.com/report/github.com/moov-io/ach-test-harness)
[![Apache 2 licensed](https://img.shields.io/badge/license-Apache2-blue.svg)](https://raw.githubusercontent.com/moov-io/ach-test-harness/master/LICENSE)

# moov-io/ach-test-harness

A configurable FTP server and Go library to interactively test ACH scenarios to replicate real world originations, returns, changes, prenotes, and transfers.

If you believe you have identified a security vulnerability please responsibly report the issue as via email to security@moov.io. Please do not post it to a public issue tracker.

## Project status

This project is used in production at an early stage and might undergo breaking changes to reach a stable status. We are looking for community feedback so please try out our code or give us feedback!

## Getting started

We publish a [public Docker image `moov/ach-test-harness`](https://hub.docker.com/r/moov/ach-test-harness/) from Docker Hub or use this repository. No configuration is required to serve FTP on `:2222` and metrics at `:3333/metrics` in Prometheus format. <!-- We also have Docker images for [OpenShift](https://quay.io/repository/moov/ach-test-harness?tab=tags) published as `quay.io/moov/ach-test-harness`. -->

The harness serves FTP only. It has no SFTP server.

### Docker image

Start the Docker image with `docker run`. Run this command from a clone of this repository, so that `./examples` is mounted with the example config. To use other versions, see the [tags on Docker Hub](https://hub.docker.com/r/moov/ach-test-harness/tags).

```
$ docker run -p 2222:2222 -p 3333:3333 -p 30000-30009:30000-30009 \
    -v $PWD/examples:/examples -e APP_CONFIG=/examples/config.yml \
    moov/ach-test-harness:latest
```

Without `-v` and `-e APP_CONFIG`, the harness starts with the default config, which has no responses.

### Docker Compose

`docker-compose.yml` in this repository starts the harness with `./examples/config.yml`. Run it from a clone of this repository:

```
$ docker compose up
harness-1  | ts=... msg="loading config file" app=ach-test-harness component=Service file=/configs/config.default.yml level=info version=...
harness-1  | ts=... msg="Loading APP_CONFIG config file" APP_CONFIG=/examples/config.yml app=ach-test-harness component=Service level=info version=...
harness-1  | ts=... msg="loading config file" APP_CONFIG=/examples/config.yml app=ach-test-harness component=Service file=/examples/config.yml level=info version=...
harness-1  | ts=... msg="creating data/outbound" app=ach-test-harness level=info version=...
harness-1  | ts=... msg="creating data/returned" app=ach-test-harness level=info version=...
harness-1  | ts=... msg="creating /reconciliation/" app=ach-test-harness level=info version=...
harness-1  | ts=... msg="listening on [::]:3333" app=ach-test-harness level=info version=...
harness-1  | ts=... msg="matcher: enable debug logging" app=ach-test-harness level=info version=...
```

If you have the older Compose v1 binary, use `docker-compose up`.

### From source

You need Go 1.26 or newer. `make install` downloads the dependencies into `vendor/`, which `make build` needs.

```
$ git clone https://github.com/moov-io/ach-test-harness.git
$ cd ach-test-harness
$ make install update build
$ APP_CONFIG=./examples/config.yml ./bin/ach-test-harness
```

You can also run `go run ./cmd/ach-test-harness` from the repository. The harness writes files under `RootPath` (`./data` in the example config). Run `./bin/ach-test-harness -h` to see the environment variables.

### Use the harness

You can then use an FTP client that connects to `localhost:2222` with a username of `admin` and password of `secret`. Upload files to the `outbound/` directory and watch for any responses in the `returned/` directory. For example, with `curl`:

```
$ curl -T examples/ppd-debit.ach ftp://admin:secret@localhost:2222/outbound/
$ curl ftp://admin:secret@localhost:2222/returned/
```

The admin server on `:3333` serves `/live`, `/ready`, `/metrics` and the [search API](docs/entry-search.md) (`/api/entries` and `/api/batches`).

### config.yml

After setup inspect the configuration file in `./examples/config.yml` and setup some scenarios to match uploaded files.

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
  Matching:
    # Enable for verbose logging of incoming entries and the matches/actions.
    Debug: false
  # ValidateOpts can use all values from https://pkg.go.dev/github.com/moov-io/ach#ValidateOpts
  ValidateOpts: {}
  Responses:
    # Entries that match both the DFIAccountNumber and TraceNumber will be returned with a R03 return code.
    # Each entry will be checked against all match conditions and actions of the first match will be used.
    - match:
        accountNumber: "12345678"
        traceNumber: "121042880000001"
      action:
        return:
          code: "R03"

    - match:
        amount:
          value: 12357 # $123.57
        action:
          delay: "12h"
          return:
            code: "R10"
```

#### config schema

The full config for Responses is below:

```yaml
# All populated fields must match for the action to be applied to the BatchHeader and EntryDetail
match:
  # Match the DFIAccountNumber on the EntryDetail
  accountNumber: <string>
  amount:
    min: <integer>
    max: <integer>
    value: <integer>       # Either min AND max OR value is used
  individualName: <string> # Compare the IndividualName on EntryDetail records
  routingNumber: <string>  # Exact match of ABA routing number (RDFIIdentification and CheckDigit)
  traceNumber: <string>    # Exact match of TraceNumber
  entryType: <string>      # Checks TransactionCode. Accepted values: credit, debit or prenote. Also can be Nacha value (e.g. 27, 32)

  # Match on BatchHeader fields
  companyIdentification: <string>
  companyEntryDescription: <string>

# The not block acts as exclusionary matches on BatchHeader and EntryDetail to prevent matching certain values.
# Any match here will prevent the action from being selected and ran.
not:
  # Match the DFIAccountNumber on the EntryDetail
  accountNumber: <string>
  amount:
    min: <integer>
    max: <integer>
    value: <integer>       # Either min AND max OR value is used
  individualName: <string> # Compare the IndividualName on EntryDetail records
  routingNumber: <string>  # Exact match of ABA routing number (RDFIIdentification and CheckDigit)
  traceNumber: <string>    # Exact match of TraceNumber
  entryType: <string>      # Checks TransactionCode. Accepted values: credit, debit or prenote. Also can be Nacha value (e.g. 27, 32)

  # Match on BatchHeader fields
  companyIdentification: <string>
  companyEntryDescription: <string>


# Matching will find at most two Actions in the config file order. One Copy Action and one Return/Correction Action.
# Both actions will be executed if the Return/Correction Action has a delay.
# Valid combinations include:
#  1. Copy
#  2. Return/Correction with Delay
#  3. Return/Correction without Delay
#  4. Copy and Return/Correction with Delay
#  5. Nothing
# Invalid combinations are:
#  1. Copy and Return/Correction without Delay
#  2. Copy with Delay (validated when reading configuration)
action:
  # How long into the future should we wait before making the correction/return available?
  delay: <duration>

  # Copy the EntryDetail to another directory (not valid with a delay)
  copy:
    path: <string> # Filepath on the FTP server

  # Send the EntryDetail back with the following ACH change code
  correction:
    code: <string>
    data: <string>

  # Send the EntryDetail back with the following ACH return code
  return:
    code: <string>
```

### Delays

`delay` is a Go duration, for example `"30s"`, `"1m"` or `"12h"`. To test delays quickly, use a short value such as `"1m"`.

- The harness writes a delayed return or correction file to the `Return` path (for example `returned/`) at once, and sets the modification time of the file to the time when the delay ends.
- FTP directory lists do not show a file until its modification time is in the past.
- The file is on disk, so a restart does not lose it. When the delay ends, the file shows in the list.

### Return and change codes

- `return.code` is a Nacha return code such as `R01` or `R03`. The harness accepts the codes that [`ach.LookupReturnCode`](https://pkg.go.dev/github.com/moov-io/ach#LookupReturnCode) knows. See [moov-io/ach returns](https://github.com/moov-io/ach/blob/master/docs/returns.md).
- `correction.code` is a Nacha change code such as `C01`, and `correction.data` is the corrected data that goes in the Addenda98 record (for example the correct account number for `C01`). See [moov-io/ach changes](https://github.com/moov-io/ach/blob/master/docs/changes.md) for the codes and the data format of each code.

## Examples

### Return debits between two values

```
  - match:
      entryType: "debit"
      amount:
        min: 100000 # $1,000
        max: 120000 # $1,200
    action:
      return:
        code: "R01"
```

### Return a specific TraceNumber

```
  - match:
      # This matches ./examples/ppd-debit.ach
      traceNumber: "121042880000001"
    action:
      return:
        code: "R03"
```

### Correct an account number
```
  - match:
      # This matches ./examples/utility-bill.ach
      accountNumber: "744-5678-99"
    action:
      correction:
        code: "C01"
        data: "744567899"
```

### Copy debit entries for a routing number
```
  - match:
      entryType: "debit"
      routingNumber: "111222337"
    action:
      copy:
        path: "/fraud-doublecheck/"
```

## Getting help

 channel | info
 ------- | -------
[Entry Search API](docs/entry-search.md) | Search the entries and batches on the FTP server with `/api/entries` and `/api/batches`.
Twitter [@moov](https://twitter.com/moov)	| You can follow Moov.io's Twitter feed to get updates on our project(s). You can also tweet us questions or just share blogs or stories.
[GitHub Issue](https://github.com/moov-io/ach-test-harness/issues) | If you are able to reproduce a problem please open a GitHub Issue under the specific project that caused the error.
[moov-io slack](https://slack.moov.io/) | Join our slack channel (`#ach`) to have an interactive discussion about the development of the project.

## Supported and tested platforms

- 64-bit Linux (Ubuntu, Debian), macOS, and Windows

## Contributing

Yes please! Please review our [Contributing guide](CONTRIBUTING.md) and [Code of Conduct](https://github.com/moov-io/ach/blob/master/CODE_OF_CONDUCT.md) to get started! Checkout our [issues for first time contributors](https://github.com/moov-io/ach-test-harness/contribute) for something to help out with.

This project uses [Go Modules](https://github.com/golang/go/wiki/Modules) and uses Go 1.26 or higher. See [Golang's install instructions](https://golang.org/doc/install) for help setting up Go. You can download the source code and we offer [tagged and released versions](https://github.com/moov-io/ach-test-harness/releases/latest) as well. We highly recommend you use a tagged release for production.

### Test coverage

Improving test coverage is a good candidate for new contributors while also allowing the project to move more quickly by reducing regressions issues that might not be caught before a release is pushed out to our users. One great way to improve coverage is by adding edge cases and different inputs to functions (or [contributing and running fuzzers](https://github.com/dvyukov/go-fuzz)).

Tests can run processes (like sqlite databases), but should only do so locally.

## License

Apache License 2.0 See [LICENSE](LICENSE) for details.
