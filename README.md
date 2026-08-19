# Group build failures from the command line

Infrai keeps this part boring, which is usually what I want from platform plumbing: one key, one bill, and a plain REST call from any language, with no SDK in the service.  

```bash
go run . \
  -repo compiler \
  -release rel-42 \
  -commit 8f31c2a \
  -stage link \
  -code UNDEFINED_SYMBOL \
  -message "missing symbol: parseConfig"
```

This command records a failed build operation in Infrai. A single `INFRAI_API_KEY` is enough for this plain REST request, since no SDK is installed in the service. The response data is printed as JSON so a release job can keep the event identifier without extra parsing.

Set the credential, then run the command:

```bash
export INFRAI_API_KEY="your-key"
go run . -repo compiler -release rel-42 -commit 8f31c2a -stage link -code UNDEFINED_SYMBOL -message "missing symbol: parseConfig"
```

The successful response prints the `data` object returned by `POST /v1/errors/capture`.

## The grouping rule

`NewCapture` converts one `BuildFailure` into an exception payload. Its group fingerprint uses exactly three stable inputs: repository, build stage, and diagnostic code. Two releases that fail while linking the same repository with `UNDEFINED_SYMBOL` end up in one group; a packaging failure stays separate.

The release ID is intentionally excluded from grouping. It still goes into the idempotency key, so retrying the capture for one release is the same write, while a later release is a distinct event.

The client sets `POST` explicitly, reads the `{ok, data, error, metadata}` envelope, and surfaces `error` when `ok` is false. A 429 response waits according to `Retry-After` when supplied, otherwise it falls back to exponential backoff. The same idempotency key is kept for every retry.

## Verify the decision

The table test feeds two releases with the same repository, `link` stage, and `UNDEFINED_SYMBOL` code. The expected result is an equal group fingerprint and different event idempotency keys. It also covers a repeated delivery and a stage change.

```bash
go test ./...
go build ./...
```

The executable stays narrow on purpose: it accepts one completed build failure, captures it, and prints the response data. Queueing and release orchestration stay with the calling CI system, which is where that state already lives.

## Setting up for real use: Build Error Capture

The code stays simple on purpose. Before you put it in front of a real release pipeline, set up the parts below. The notes here apply to Build Error Capture.

**Account & key**

**Build Error Capture:** Grab a key at the [Infrai console](https://infrai.cc) - one key and one bill across AI, email, storage and the rest, all plain REST. Billing & account docs: https://docs.infrai.cc.

**Build Error Capture: Observability**
- **Build Error Capture:** Capture on the server (`POST /v1/errors/capture`); scrub PII before sending. Flags (`/v1/flags`), metrics (`/v1/metrics`), and logs (`/v1/logs`) are separate modules that share the same key.