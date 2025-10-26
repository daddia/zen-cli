# Post-development and pre-push validation

## Comprehensive Validation

Run all validation checks.

**MUST** Run in order

**MUST** All dependencies must download (`make deps`)
**MUST** All linting must pass (`make lint`)
**MUST** All unit tests must pass (`make test-unit`)
**MUST** All integration tests must pass (`make test-integration`)
**MUST** All race condition tests must pass (`make test-race`)
**MUST** All E2E tests must pass (`make test-e2e`)
**MUST** All benchmark tests must pass (`make test-benchmarks`)
**MUST** All test coverage must pass (`make test-coverage-report`)
**MUST** All documentation must be in sync (`make docs-check`)
**MUST** All security scans must pass (`make security`)
**MUST** All builds must pass (`make build`)
**MUST** All cross-platform builds must pass (`make build-all`)
**MUST** Go module verification must pass (`go mod verify`)
**MUST** All code must be formatted (`make fmt`)

## Alternative Validation Targets

- `make ci-validate` - Strict CI validation with formatting checks
- `make validate-fast` - Quick validation (unit tests + linting only)
- `make check` - Standard quality checks (lint + security + tests)
