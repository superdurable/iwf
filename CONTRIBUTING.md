# Contributing

## Prerequisites

- Go 1.24+ (see `server/go.mod`; root `go.work` pins the workspace)
- Java 8+ and Gradle wrapper (for `sdk-java` / `samples-java`)
- Python 3.9+ and Poetry (for `sdk-python` / `samples-python`)
- Docker (for integration tests / local Temporal+Cadence stacks)

## Go workspace

Root [`go.work`](go.work) includes `sdk-go` and `samples-go` (local `replace` for the SDK).

```bash
go work sync
go build ./sdk-go/...
go build ./samples-go/...
```

### Server

Build outside the workspace (or with `GOWORK=off`) so Cadence’s older `genproto` does not clash with Temporal’s split modules:

```bash
cd server && go build ./...
make -C server unitTests
# Integration tests need Cadence/Temporal; see server/CONTRIBUTING.md
```

### Go SDK

```bash
make -C sdk-go ci-tests   # may start docker compose under sdk-go/integ
```

## IDL / OpenAPI (`protos/`)

Specs live in [`protos/`](protos/) (`iwf.yaml` for server, `iwf-sdk.yaml` for SDKs). There is no git submodule.

Regenerate after editing specs:

```bash
make -C server idl-code-gen
make -C sdk-go idl-code-gen
# Java: Gradle OpenAPI generate tasks in sdk-java/
# Python: see sdk-python/README.md (openapi-python-client)
```

## Java

```bash
cd sdk-java && ./gradlew build
cd ../samples-java && ./gradlew build
```

## Python

```bash
cd sdk-python && poetry install && poetry run pytest   # if tests are configured
cd ../samples-python && poetry install
```

## CI

Root workflows under [`.github/workflows/`](.github/workflows/) run path-filtered jobs for server and each SDK/samples tree. Prefer fixing those over re-adding nested `*/.github/workflows` duplicates.

## Package-specific guides

- [server/CONTRIBUTING.md](server/CONTRIBUTING.md)
- [sdk-go/CONTRIBUTION.md](sdk-go/CONTRIBUTION.md)
- [sdk-java/README.md](sdk-java/README.md)
- [sdk-python/README.md](sdk-python/README.md)
