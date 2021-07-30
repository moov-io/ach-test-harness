<!-- generated-from:d6ef15f91b0edc74c8eeb3ea0138b6bad74e43fe69d00204b961f761948aa5b9 DO NOT REMOVE, DO UPDATE -->
# ACH Test Harness
**Purpose** | **[Configuration](CONFIGURATION.md)** | **[Running](RUNNING.md)** | **[Client](../pkg/client/README.md)**

---

## Purpose

A configurable FTP/SFTP server and Go library to interactively test ACH scenarios to replicate real world originations, returns, changes, prenotes, and transfers.

## Search

ach-test-harness offers search over the files, batches, and entries on the underlying filesystem. This is useful for automated testing as well as dashboards when used as a sandbox environment.

### Entries

```
GET /entries?traceNumber=YYYYY
```

This endpoint will return entries matching the query params provided. The logic is similar to Response matching with the FTP interface. The supported query params are:

- `accountNumber=YYYY` returns entries with matching `DFIAccountNumber` values
- `amount=YYYY` returns entries with matching `Amount` values
- `routingNumber=YYYY` returns entries with matching `RDFIIdentification` values
- `traceNumber=YYYY` returns entries with matching `TraceNumber` values
- `createdAfter=YYYY` returns entries from files created after the timestamp (in `FileCreationDate` and `FileCreationTime`)
   - Supported timestamp values are:
      - ISO8601 (`2018-11-18T09:04:23-08:00`)
      - YYYY-MM-DD (`2021-07-21`)
      - RFC3339 (`2006-01-02T15:04:05Z07:00`)

This endpoint will return the following response:

```json
[
  {
    "id":"",
    "transactionCode":27,
    "RDFIIdentification":"23138010",
    "checkDigit":"4",
    "DFIAccountNumber":"744-5678-99      ",
    "amount":500000,
    "identificationNumber":"location1234567",
    "individualName":"Best Co. #123456789012",
    "discretionaryData":"S ",
    "traceNumber":"031300010000001"
  },
  {
    "id":"",
    "transactionCode":27,
    "RDFIIdentification":"23138010",
    "checkDigit":"4",
    "DFIAccountNumber":"744-5678-99      ",
    "amount":125,
    "identificationNumber":"Fee123456789012",
    "individualName":"Best Co. #123456789012",
    "discretionaryData":"S ",
    "traceNumber":"031300010000002"
  },
  {
    "id":"",
    "transactionCode":22,
    "RDFIIdentification":"23138010",
    "checkDigit":"4",
    "DFIAccountNumber":"987654321        ",
    "amount":100000000,
    "identificationNumber":"               ",
    "individualName":"Credit Account 1      ",
    "discretionaryData":"  ",
    "traceNumber":"121042880000002"
  }
]
```

## Getting help

 channel | info
 ------- | -------
 [Project Documentation](https://github.com/moov-io/ach-test-harness/tree/master/docs/) | Our project documentation available online.
Twitter [@moov](https://twitter.com/moov)	| You can follow Moov.io's Twitter feed to get updates on our project(s). You can also tweet us questions or just share blogs or stories.
[GitHub Issue](https://github.com/moov-io/ach-test-harness/issues) | If you are able to reproduce a problem please open a GitHub Issue under the specific project that caused the error.
[moov slack](https://slack.moov.io/) | Join our slack channel (`#ach-test-harness`) to have an interactive discussion about the development of the project.

---
**[Next - Configuration](CONFIGURATION.md)**
