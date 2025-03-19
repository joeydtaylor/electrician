# ⚙️ Internal Directory

The **internal** directory contains the **foundational components** of Electrician. These packages **power the framework’s core functionality**, enabling structured concurrency, event-driven pipelines, data transformation, and microservice communication.

These packages are **not** intended for direct use by end users. Instead, they provide **infrastructure** that Electrician’s higher-level APIs (such as `pkg/builder`) use to construct pipelines, enforce reliability mechanisms, and manage performance.

---

## 📦 Package Overview

Each package within `internal/` has a **specific role** in the Electrician ecosystem. Below is a breakdown of all internal packages and their responsibilities.

| Package             | Description                                                                                                 |
| ------------------- | ----------------------------------------------------------------------------------------------------------- |
| **adapter/**        | Connects external systems via **HTTP, OAuth2, or other data sources**.                                      |
| **circuitbreaker/** | Implements **failure detection and recovery**, including automatic reset logic.                             |
| **codec/**          | Handles **serialization and deserialization** of data within pipelines.                                     |
| **conduit/**        | Manages **multi-stage processing** by linking multiple Wires together.                                      |
| **forwardrelay/**   | **Sends** data streams to external services via gRPC (outgoing messages).                                   |
| **generator/**      | Produces **autonomous data feeds** for a pipeline, such as synthetic test data.                             |
| **internallogger/** | Centralized **structured logging** for internal components (**only non-standard dependency: `zap`**).       |
| **meter/**          | Measures **performance metrics**, tracking throughput, errors, and concurrency statistics.                  |
| **plug/**           | Defines **pluggable data sources** (like adapters), allowing dynamic integration of new input sources.      |
| **receivingrelay/** | **Receives** and decodes incoming gRPC streams, integrating external data into pipelines.                   |
| **relay/**          | Contains **compiled Protobuf gRPC** code for Electrician’s inter-service communication (**uses protobuf**). |
| **resister/**       | Implements **queue-based rate limiting** for controlled message delivery.                                   |
| **sensor/**         | Monitors system behavior, collecting **real-time telemetry** for debugging and analytics.                   |
| **surgeprotector/** | Protects pipelines from **overload**, rerouting or pausing traffic when thresholds are exceeded.            |
| **types/**          | **Source of truth** for all shared interfaces and contracts in Electrician.                                 |
| **utils/**          | Contains **internal utilities** (e.g., functors, common helpers).                                           |
| **wire/**           | The **core processing unit** of Electrician, handling data flow, transformation, and lifecycle management.  |

---

## 📂 Package Structure

Each internal package follows a **consistent structure** designed for **maintainability and scalability**.

| File Name           | Purpose                                                          |
| ------------------- | ---------------------------------------------------------------- |
| **api.go**          | Public API for this package (**internal to Electrician only**).  |
| **internal.go**     | Private helper functions, **only used within the package**.      |
| **notify.go**       | Handles **logging, telemetry, and event notifications**.         |
| **options.go**      | Provides **functional options** for configuring components.      |
| **package.go**      | Defines the **type and constructor** for the component.          |
| **package_test.go** | Unit/integration tests ensuring **correctness and reliability**. |

### ✨ **Why This Structure?**

✔️ **Separation of Concerns** – Public API (`api.go`) is distinct from internal logic (`internal.go`).  
✔️ **Composability** – Configuration via **functional options (`options.go`)** or direct method calls.  
✔️ **Event-Driven Architecture** – `notify.go` ensures **consistent logging and telemetry** across the framework.  
✔️ **Reliability** – Every package includes **unit tests (`package_test.go`)** for guaranteed stability.

---

## 🚀 Design Philosophy

Electrician follows a **strict immutability model**:  
🔴 **Runtime modifications to components are not supported.**  
✔️ **Pipelines should be fully configured before execution.**

This ensures:

- **Predictability** – Pipelines behave consistently without unexpected state changes.
- **Concurrency Safety** – Eliminates race conditions and synchronization issues.
- **Performance Optimization** – Avoids costly reconfiguration overhead.

While Go technically allows modifying running components, **Electrician strongly discourages this**. The framework is designed for **deterministic, event-driven pipelines** that are pre-configured and stable.

---

## 🔧 Extending Internal Packages

To **add new functionality**, follow this **strict contract**:

1️⃣ **Modify `types/`**

- All public functionality starts by **adding an interface method** to `types/package.go`.
- This ensures **consistent contracts** across components.

2️⃣ **Implement in `api.go`**

- Add the **actual implementation** of the new method.

3️⃣ **Create a Functional Option in `options.go`**

- Supports **composable, declarative-style configuration**.

4️⃣ **Enhance `notify.go` (if applicable)**

- If your change introduces **new events**, add corresponding **logging and telemetry hooks**.

5️⃣ **Update Tests (`package_test.go`)**

- Ensure **full test coverage** before merging.

By enforcing these steps, Electrician maintains **consistency, safety, and extensibility**.

---

## 🏗️ Relationship to `pkg/builder`

- **End users never interact with `internal/` directly**—they use `pkg/builder`, which acts as the **public API**.
- **The `builder/` package is a wrapper** that applies functional options and calls the appropriate internal components.
- The **internal directory is effectively the "engine" of Electrician**, while `builder/` provides a user-friendly interface.

---

## ⚡ Standard Library First

Electrician is **99% based on the Go standard library**.  
The **only** external dependencies used in `internal/` are:

- **`zap` (Uber’s logging library)** – Used in `internallogger/`, as it’s the **fastest and most structured** logging system in Go.
- **`protobuf` (gRPC and serialization)** – Used in `relay/`, since Electrician supports **cross-service messaging**.

Everything else—including **networking, data transformation, and concurrency management**—is built using **pure Go**, ensuring:

✅ **Maximum compatibility** – No unnecessary dependencies.  
✅ **Minimal attack surface** – Secure and easy to audit.  
✅ **High performance** – Optimized for **low-latency, high-throughput pipelines**.

---

## 📖 Further Reading

- **[Root README](../../../README.md)** – Electrician’s overall architecture and principles.
- **[Per-Package README]** – Each internal sub-package (`wire`, `circuitbreaker`, etc.) contains additional details.
- **[Examples Directory](../../../example/)** – Demonstrates these internal components in real-world use cases.

---

## 📝 License

All internal packages fall under Electrician’s [Apache 2.0 License](../../../LICENSE).  
You are free to use, modify, and distribute them under these terms, but note that **they are not designed for direct external use**.

---

## ⚡ Happy wiring! 🚀

If you have any questions or need clarification, feel free to **open a GitHub issue**.
