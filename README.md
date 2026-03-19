# Shipment Tracking gRPC Microservice

This project implements a small shipment tracking microservice for the backend test task from `vektor-backend-test-task.pdf`.

## What The Service Does

The service allows clients to:

- create a shipment
- retrieve shipment details
- append shipment status events
- fetch shipment event history

The current shipment status is derived from valid lifecycle transitions and is kept consistent with persisted event history.

## Architecture Overview

The project follows a layered structure:

- `internal/domain`
  Contains core business rules, shipment entity, statuses, validation, and repository abstraction.
- `internal/app`
  Contains use-case orchestration for creating shipments and appending status events.
- `internal/infrastructure`
  Contains SQLite persistence and transactional data access.
- `internal/transport/grpc`
  Contains the gRPC adapter and request-to-domain error mapping.
- `proto`
  Contains the protobuf service contract.
- `gen`
  Contains generated Go files for protobuf and gRPC bindings.

This keeps business rules independent from gRPC and SQLite details, which makes the core logic testable in isolation.

## Design Decisions

- Shipment lifecycle is modeled in the domain layer.
  Valid transitions are:
  `PENDING -> PICKED_UP -> IN_TRANSIT -> DELIVERED`
- Duplicate status updates are rejected.
- Shipment validation is performed before persistence.
- Shipment creation and status update flows persist shipment state and event history atomically in the repository layer.
- `driver_details` stores both driver information and assigned unit/vehicle information in one field to keep the contract compact.

## Assumptions

- `reference_number` is unique per shipment.
- `amount` must be greater than zero.
- `driver_revenue` must be zero or positive and cannot exceed `amount`.
- Empty origin, destination, reference number, and driver/unit details are invalid.
- The protobuf contract currently exposes the main tracking lifecycle used by the task.

## Run The Service

PowerShell:

```powershell
$env:PORT="50051"
$env:DB_DSN="./shipments.db"
go run ./cmd
```

The server starts on `localhost:50051` by default.

## Run Tests

In this environment, the default Go build cache path may be restricted, so use a local cache:

```powershell
$env:GOCACHE="$PWD\\.gocache"
go test ./...
```

## Build

```powershell
$env:GOCACHE="$PWD\\.gocache"
go build -o .\bin\server.exe ./cmd
```

## Protobuf

If `protoc`, `protoc-gen-go`, and `protoc-gen-go-grpc` are installed locally, regenerate the bindings with:

```powershell
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/shipment.proto
```

## Notes

- SQLite was chosen to keep the task self-contained and easy to run locally.
- The lifecycle is intentionally kept compact and focused on the main shipment tracking flow requested in the assignment.
