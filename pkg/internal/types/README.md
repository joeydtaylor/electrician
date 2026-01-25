# types

The types package defines shared interfaces and constants used across Electrician's internal components. It is the contract layer that keeps components interoperable without tight coupling.

## Responsibilities

- Define component interfaces (Wire, Plug, Generator, Sensor, Meter, etc.).
- Provide shared constants for metrics, log levels, and component metadata.
- Establish common configuration and option patterns.

## Design notes

- Interfaces here are used across multiple internal packages.
- Types that are only used by a single package should live with that package.
- This package does not implement behavior; it is intentionally lightweight.

## Key files

- meter.go: metric names, MetricInfo, Meter interface
- wire.go, plug.go, generator.go: component contracts
- logger.go: logging contract and levels
- common.go: shared option types and metadata

## Usage

Internal packages implement these interfaces; the builder package returns concrete implementations that satisfy them. This keeps the public API stable while internals can evolve.

## License

Apache 2.0. See ../../../LICENSE.
