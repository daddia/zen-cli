# Post-development and pre-push validation

**MUST** All CI must pass
**MUST** All dependencies must download (`make deps`)
**MUST** All code must be formatted (`make fmt`)
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
