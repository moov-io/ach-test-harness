## Entry Search

ach-test-harness offers search over the files, batches, and entries on the underlying filesystem. This is useful for automated testing as well as dashboards when used as a sandbox environment.

The search endpoints are on the admin server (`Servers.Admin.Bind.Address`, `:3333` by default) under the `/api` prefix:

```
GET http://localhost:3333/api/entries?traceNumber=YYYYY
GET http://localhost:3333/api/batches?traceNumber=YYYYY
```

`/api/entries` returns the matching entries. `/api/batches` returns the batches that contain matching entries, with their batch header, the matching entries and the batch control. The logic is similar to Response matching with the FTP interface. The supported query params are:

- `accountNumber=YYYY` returns entries with matching `DFIAccountNumber` values
- `amount=YYYY` returns entries with matching `Amount` values (`/api/entries` only)
- `routingNumber=YYYY` returns entries with matching `RDFIIdentification` values
- `traceNumber=YYYY` returns entries with matching `TraceNumber` values
- `createdAfter=YYYY` returns entries from files created after the timestamp (in `FileCreationDate` and `FileCreationTime`)
   - Supported timestamp values are:
      - ISO8601 (`2018-11-18T09:04:23-08:00`)
      - YYYY-MM-DD (`2021-07-21`)
      - RFC3339 (`2006-01-02T15:04:05Z07:00`)
- `path=YYYY` searches only the files under this directory, relative to the FTP `RootPath` (for example `path=outbound`). Without `path`, the search includes all files under `RootPath`, also the files that the harness wrote (for example in `returned/`).

### Examples

After you upload `examples/ppd-debit.ach` to `outbound/`:

```
$ curl "http://localhost:3333/api/entries?traceNumber=121042880000001&path=outbound"
```

```json
[
  {
    "id": "",
    "transactionCode": 27,
    "RDFIIdentification": "23138010",
    "checkDigit": "4",
    "DFIAccountNumber": "12345678",
    "amount": 100000000,
    "identificationNumber": "               ",
    "individualName": "Receiver Account Name ",
    "discretionaryData": "  ",
    "addendaRecordIndicator": 0,
    "traceNumber": "121042880000001",
    "category": "Forward",
    "lineNumber": 3
  }
]
```

```
$ curl "http://localhost:3333/api/batches?traceNumber=121042880000001&path=outbound"
```

```json
[
  {
    "batchHeader": {
      "id": "",
      "serviceClassCode": 225,
      "companyName": "Name on Account",
      "companyIdentification": "121042882",
      "standardEntryClassCode": "PPD",
      "companyEntryDescription": "REG.SALARY",
      "effectiveEntryDate": "190625",
      "settlementDate": "   ",
      "originatorStatusCode": 1,
      "ODFIIdentification": "12104288",
      "batchNumber": 1,
      "lineNumber": 2
    },
    "entryDetails": [
      {
        "id": "",
        "transactionCode": 27,
        "RDFIIdentification": "23138010",
        "checkDigit": "4",
        "DFIAccountNumber": "12345678",
        "amount": 100000000,
        "identificationNumber": "               ",
        "individualName": "Receiver Account Name ",
        "discretionaryData": "  ",
        "addendaRecordIndicator": 0,
        "traceNumber": "121042880000001",
        "category": "Forward",
        "lineNumber": 3
      }
    ],
    "batchControl": {
      "id": "",
      "serviceClassCode": 225,
      "entryAddendaCount": 1,
      "entryHash": 23138010,
      "totalDebit": 100000000,
      "totalCredit": 0,
      "companyIdentification": "121042882",
      "ODFIIdentification": "12104288",
      "batchNumber": 1,
      "lineNumber": 4
    },
    "offset": null
  }
]
```

An error returns status 500 with a JSON body, for example `{"error":"invalid path"}` when `path` is absolute or contains `..`.
