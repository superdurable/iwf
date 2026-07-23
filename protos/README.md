# iWF IDL (`protos/`)

Protobuf + gRPC interface between iWF SDKs and the iWF server.

- Source: [`iwf.proto`](iwf.proto)
- Rename catalog: [`../docs/design/idl-renames.md`](../docs/design/idl-renames.md)
- License: MIT ([`LICENSE`](LICENSE))

## Services

- **FlowService** — hosted by the server; SDKs call these RPCs
- **WorkerService** — hosted by the worker; the server calls `WaitFor`, `Execute`, and `WorkerRpc`

## Codegen

Regenerate checked-in stubs into server and SDK trees:

```bash
make -C protos proto
```

| Output | Replaces |
|--------|----------|
| `server/gen/iwfpb/` | `server/gen/iwfidl/` |
| `sdk-go/gen/iwfpb/` | `sdk-go/gen/iwfidl/` |
| `sdk-java/src/main/java/io/superdurable/gen/` | OpenAPI `build/generated` |
| `sdk-python/iwf/iwfpb/` | `sdk-python/iwf/iwf_api/` |

`make -C server idl-code-gen` and `make -C sdk-go idl-code-gen` delegate to `make -C protos proto`.
