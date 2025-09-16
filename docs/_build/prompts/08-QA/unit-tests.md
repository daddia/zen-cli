# Prompt: Generate Comprehensive Test Suite

<!-- Recommend using o3-mini (or similar fast advanced reasoning model) -->

<!-- Update and copy prompt from here -->

Conduct a comprehensive review of the codebase.

Review the shared context and configurations to identify any existing tests, tests framework and libraries in use.

**Your task** is to generate a comprehensive suite of unit tests for files in `src/` directory. 

---

**Language:** Typescript v5.8.3
**Test Frameworks:** vitest v3.1.4
**Test Libraries:** @testing-library/react MSW
**Test Coverage:** istanbul

Ensure librariies are correctly installed.

```sh
pnpm i -Dw vitest @testing-library/react msw nyc
```

---

## **Test Directory Structure**

* All tests MUST BE organised in centralised `tests/` directory structure
* Mirror `src/` under `unit/` and `integration/`
* Separate test types (unit, integration, e2e, smoke) to control execution scope and speed
* Keep static data in `fixtures/`
* Keep mocks in `mocks/`
* Centralise setup in `setup/` to ensure runner config stays clean and consistent across all tests

Below is our centralised `tests/` directory structure.

```
tests/
├── unit/                           # Isolated, fast-running tests
│   ├── components/                 # React component tests
│   │   ├── Button.test.tsx
│   │   └── ...
│   ├── hooks/                      # Custom hook tests
│   │   └── useAuth.test.ts
│   ├── utils/                      # Pure function tests
│   │   └── formatDate.test.ts
│   └── services/                   # API-wrapper logic tests
│       └── apiClient.test.ts
│
├── integration/                    # Tests exercising multiple modules
├── e2e/                            # End-to-end browser flow tests
├── smoke/                          # Lightweight sanity checks

├── fixtures/                       # Static test data (JSON, HTML, snapshots)
│   ├── users.json
│   └── ...
│
├── mocks/                          # Custom HTTP or module mocks
│   ├── server.ts                   # MSW server setup
│   └── fetchMock.ts
│
├── helpers/                        # Reusable test utilities & custom matchers
│   └── renderWithProviders.tsx
│
└── setup/                          # Global setup & teardown code
    ├── vitest.setup.ts             # Vitest config, global mocks, a11y rules
    └── teardown.ts                 # Cleanup after all tests complete
```

---

## **RULES FOR UNIT TESTS**

All unit tests MUST follow these guidelines:

* **Coverage:** Unit tests **SHALL** achieve overall coverage of at least >60 %.
* **Isolate Tests:** Unit tests **MUST** be independent and **SHALL NOT** share state.
* **Fast Execution:** Unit tests **SHOULD** execute rapidly to encourage frequent runs, especially during development.
* **Clear Naming:** Unit tests **SHOULD** use descriptive names that state the expected behaviour.
* **Single Responsibility:** Unit tests **SHALL** verify exactly one behaviour or aspect per test.
* **Setup & Teardown:** Setup and teardown hooks **SHOULD** be used to initialise test state and clean up after each test.
* **Assertions:** Unit tests **MUST** use explicit, assertive assertions that fail fast and clearly.
* **Mocking:** Unit tests **MUST** mock or stub all external dependencies to isolate the unit under test.
* **Edge Cases:** Unit tests **SHOULD** cover typical cases, edge conditions, and failure modes.
* **Readability:** Unit tests **SHOULD** be written for clarity and ease of understanding.
* **Helpers:** Unit tests **MAY** employ helper functions to abstract complex setup or repetitive logic.
* **CI Integration:** Unit tests **SHALL** be automated in the CI pipeline to catch regressions early.
* **Documentation:** Comments in unit tests **ARE OPTIONAL** and **SHOULD** be limited to non-obvious or complex scenarios.
* **Test Structure:** Unit tests **MUST** follow the Arrange-Act-Assert pattern.
* **Exception Testing:** Unit tests **SHOULD** include assertions for expected exceptions to validate error handling.
* **Correctness:** Unit tests **SHALL** validate correct behaviour under varying inputs and states.

---
