# Contributing

Thanks for your interest in Electrician. This project favors clear, predictable code and stable behavior under concurrency.

## Development workflow

1) Create a branch for your change.
2) Keep changes focused and scoped.
3) Update or add tests for any behavior change.
4) Run tests locally:

```bash
go test ./...
go test ./... -race
```

If a package requires external services (Kafka, S3, etc.), add tests that can run without those services or clearly document how to run the integration tests.

## Code style

- Prefer small, focused files with clear names.
- Use functional options for configuration.
- Avoid mutating component configuration after Start(); it is not supported.
- Keep hot paths allocation-light; move heavier work to setup or monitor loops.
- Add doc comments for exported types and functions.

## Documentation

- Keep README files professional and concise.
- Update examples when behavior changes.
- If you add a new package, include a README and a small example if possible.

## Questions

Open a GitHub issue for discussion, design questions, or feature proposals.
