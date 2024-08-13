## v0.13.1 (Released 2024-08-13)

IMPROVEMENTS

- response: remove noisy match spans

## v0.13.0 (Released 2024-08-12)

IMPROVEMENTS

- meta: add telemetry via OpenTracing

BUILD

- fix(deps): update module github.com/moov-io/ach to v1.41.0

## v0.12.1 (Released 2024-07-19)

IMPROVEMENTS

- service: switch logging to DiscardLogger

## v0.12.0 (Released 2024-07-18)

IMPROVEMENTS

- feat: simplify green path logging

BUILD

- fix(deps): update module github.com/moov-io/base to v0.51.1
- fix(deps): update module github.com/moov-io/ach to v1.40.1

## v0.11.0 (Released 2024-06-26)

IMPROVEMENTS

- feat: batches endpoint for searching files and returning batches

BUILD

- fix(deps): update module github.com/moov-io/ach to v1.40.3
- fix(deps): update module github.com/moov-io/base to v0.50.0

## v0.10.5 (Released 2024-05-10)

IMPROVEMENTS

- response: fix ATX/CTX AddendaCount in entry transformer

BUILD

- fix(deps): update module github.com/moov-io/ach to v1.39.2
- fix(deps): update module github.com/moov-io/base to v0.49.2

## v0.10.4 (Released 2024-04-11)

IMPROVEMENTS

- response: fixup ODFI and RDFI Identification on corrections and returns
- test: verify ODFIIdentification in return/correction BatchHeader
- test: verify every prenote correction and return transaction code

BUILD

- chore(deps): update dependency go to v1.22.2
- fix(deps): update module github.com/moov-io/ach to v1.37.2

## v0.10.3 (Released 2024-03-20)

IMPROVEMENTS

- fix: adjust TransactionCode code for prenote returns

BUILD

- chore(deps): update dependency go to v1.22.1
- fix(deps): update module github.com/moov-io/ach to v1.36.1
- fix(deps): update module github.com/stretchr/testify to v1.9.0

## v0.10.2 (Released 2024-02-23)

IMPROVEMENTS

- fix: use RDFI in batch header for returns/corrections/etc

## v0.10.1 (Released 2024-02-12)

BUILD

- build: fix issue when releasing on ARM macos

## v0.10.0 (Released 2024-02-08)

IMPROVEMENTS

- fix: cleanup matching logs, one log per entry

BUILD

- fix(deps): update module github.com/moov-io/ach to v1.34.2
- fix(deps): update module github.com/moov-io/base to v0.48.5
- test: run on macos 13.x (Intel) and 14.x (M1/ARM)
- chore(deps): update golang docker tag to v1.22

## v0.9.1 (Released 2023-11-27)

IMPROVEMENTS

- achx: fix ABA8 panic, generate correction/return trace numbers from RDFIIdentification
- fix: use fmt.Errorf instead of pkg/errors
- response: set OriginalDFI correctly on corrections/returns

## v0.9.0 (Released 2023-11-21)

IMPROVEMENTS

- response: compare entryType directly against TransactionCode as well

BUILD

- chore(deps): update dependency go to v1.21.4
- fix(deps): update module github.com/gorilla/mux to v1.8.1
- fix(deps): update module github.com/moov-io/ach to v1.33.3
- fix(deps): update module github.com/moov-io/base to v0.48.2

## v0.8.3.1 (Released 2023-09-22)

IMPROVEMENTS

- feat: generate multiple batches in one reconciliation file

BUILD

- fix(deps): update module github.com/moov-io/ach to v1.32.2
- fix(deps): update module github.com/moov-io/base to v0.46.0
- fix(deps): update go version to v1.21.0

## v0.8.2 (Released 2023-08-11)

IMPROVEMENTS

- feat: Add directory to search options on entries search API
- fix: Fix logic in transforming cases of multiple entries

BUILD

- fix(deps): update module github.com/moov-io/ach to v1.32.1
- fix(deps): update golang docker tag to v1.21

## v0.8.1 (Released 2023-08-03)

This release of ach-test-harness includes a new property (`delay: <duration>`) on actions,
which allows for returns and corrections to be produced after the initial upload. The delay
feature allows for more testing scenarios.

IMPROVEMENTS

- feat: add support for future-dated actions
- test: verify OriginalTrace is set properl

BUILD

- fix(deps): update module github.com/moov-io/ach to v1.32.0
- fix(deps): update module github.com/moov-io/base to v0.45.1

## v0.7.0 (Released 2023-02-21)

IMPROVEMENTS

- feat: add prenote entryType matcher, properly correct/return prenotes
- fix: FINAL match log
- test: linter fixup, increase coverage requirement

BUILD

- build: update moov-io/ach to v1.29.0
- chore(deps): update golang docker tag to v1.20
- fix(deps): update module github.com/moov-io/base to v0.39.0

## v0.6.9 (Released 2023-01-13)

Note: moov-io/ach version v1.28.0 does not preserve spaces in fields like `DFIAccountNumber`. Enable `PreserveSpaces: true` to restore this behavior.

BUILD

- fix(deps): update module github.com/moov-io/ach to v1.28.0
- fix(deps): update module github.com/moov-io/base to v0.38.1

## v0.6.8 (Released 2022-12-08)

BUILD

- fix(deps): update module github.com/moov-io/ach to v1.26.0
- fix(deps): update module github.com/moov-io/base to v0.37.0

## v0.6.7 (Released 2022-11-15)

BUILD

- fix(deps): update module github.com/moov-io/ach to v1.23.1
- fix(deps): update module github.com/moov-io/base to v0.36.2
- fix(deps): update module github.com/stretchr/testify to v1.8.1

## v0.6.6 (Released 2022-09-21)

IMPROVEMENTS

- response/match: use structured logging

BUILD

- build: Use go1.19.1 in CI and releases
- fix(deps): update module github.com/moov-io/ach to v1.20.1

## v0.6.4 (Released 2022-09-19)

BUILD

- build: require go 1.19.1 in CI/CD

## v0.6.3 (Released 2022-09-19)

BUILD

- chore(deps): update dependency golang to v1.19
- fix(deps): update module github.com/stretchr/testify to v1.8.0
- fix(deps): update module github.com/moov-io/base to v0.35.0

## v0.6.2 (Released 2022-05-18)

BUILD

- chore(deps): update dependency golang to v1.18
- fix(deps): update module github.com/moov-io/ach to v1.15.1
- fix(deps): update module github.com/moov-io/base to v0.29.2
- fix(deps): update module github.com/stretchr/testify to v1.7.1

## v0.6.1 (Released 2021-11-16)

IMPROVEMENTS

- entries: search even partial files

## v0.6.0 (Released 2021-11-08)

BREAKING CHANGES

moov-io/base introduces errors when unexpected configuration attributes are found in the files parsed on startup.

BUILD

- chore(deps): update moov/ach-test-harness docker tag to v0.5.2
- fix(deps): update module github.com/moov-io/ach to v1.12.2
- fix(deps): update module github.com/moov-io/base to v0.27.0

## v0.5.2 (Released 2021-09-14)

IMPROVEMENTS

- service: create Action.Copy paths on startup

BUILD

- chore(deps): update golang docker tag to v1.17
- fix(deps): update module github.com/moov-io/ach to v1.12.1
- fix(deps): update module github.com/moov-io/base to v0.24.0

## v0.5.1 (Released 2021-08-11)

BUG FIXES

- response: write parent directories if needed

BUILD

- build: enable gosec linter
- chore(deps): update moov/ach-test-harness docker tag to v0.5.0
- fix(deps): update module github.com/moov-io/base to v0.22.0

## v0.5.0 (Released 2021-07-30)

ADDITIONS

- entries: add search endpoint over account numbers, trace numbers, amount, and created at timestamps
- response: Allow setting `ach.ValidateOpts` in config

BUILD

- build: use debian stable's slim image
- fix(deps): update module github.com/moov-io/ach to v1.10.1
- fix(deps): update module github.com/moov-io/base to v0.21.1

## v0.4.1 (Released 2021-06-28)

This release contains MacOS and Windows binaries.

## v0.4.0 (Released 2021-03-30)

BREAKING CHANGES

- response: add entry type matcher for debit or credit entry

IMPROVEMENTS

- build: update moov-io/base together with gogo/protobug to fix CVE
- docs: refresh the readme after newer matchers

## v0.3.0 (Released 2021-03-22)

ADDITIONS

- response: add a "Copy" matcher for mirroring entries to another file
- response: introduce RoutingNumber matcher
- service: add route to render merged config

IMPROVEMENTS

- config: add example of routingNumber -> copy response
- fix(deps): update module github.com/moov-io/ach to v1.6.3

BUG FIXES

- configs: fix default with empty object

## v0.2.2 (Released 2021-03-10)

BUG FIXES

- response/matcher: fix strict match on amounts
- response/matcher: fix sprintf logs in matcher

IMPROVEMENTS

- response: include debug logging in matching

## v0.2.1 (Released 2021-03-08)

BUG FIXES

- response: count positive/negative matches for complex Action selection

## v0.2.0 (Released 2021-03-03)

ADDITIONS

- response: include matcher for IndividualName

IMPROVEMENTS

- response: add Matcher tests
- response: add MorphEntry transform tests

## v0.1.2 (Released 2021-03-02)

BUG FIXES

- response: don't create Batch or File if it's empty

BUILD

- fix(deps): update module github.com/moov-io/base to v0.17.0

## v0.1.1 (Released 2021-03-02)

BUILD

- Fix `docker push` command in release Action

## v0.1.0 (Released 2021-03-02)

This is the initial release of ach-test-harness. Please try it out and let us know in our `#ach` slack channel your thoughts, bugs, improvements, and feedback!
