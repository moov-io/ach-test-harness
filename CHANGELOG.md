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
