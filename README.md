# Agrirouter Go SDK

Go SDK for agrirouter is a library that provides a shared interface to access functionality common for agrirouter applications.

🚧🚧🚧 **Currently this is ONLY intended to be used internally by agrirouter outbound integrations!** 🚧🚧🚧

## Development

Following instructions are for developers working on this SDK.

### Prerequisites

- Go 1.23 or later
- bash or compatible shell
- Docker
- make (optional, for convenience)

#### Install git hooks

Run this command:

```bash
git config core.hooksPath .githooks
```

This will set up git hooks to run minimal set of checks on commit and push. These would not require `make`.

### Code Generation

Script `tools/oapi/generate.sh` would generate Go code from the OpenAPI specification files.

It uses `oapi-codegen` tool using dockerized installation, hence it only requires Docker to be installed and bash to run the script that invokes docker.

### Cheatsheet

`make` will run everything: code generation, tidying, full linting and tests.

`make test` will run tests only.

`make vet` will run `go vet` on the library code.

`make vet-test-server` will run `go vet` on the test server code.

`make vet-all` will run `go vet` on both above.

`make lint` will run vet for all library and other linters for recently added changes (against the HEAD).

`make lint-all` will run vet and linters for all files.

`make generate` will run code generation.

`make tidy` will run `go mod tidy`.

`make tidy-test-server` will run `go mod tidy` for the test server code.

`make tidy-all` will run `go mod tidy` for both library and test server code.

