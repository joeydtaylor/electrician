# ⚙️ Types Package

The **types** package is the **source of truth** for all shared interfaces and fundamental data structures in Electrician.  
Every core component—including **Wires, Circuit Breakers, Sensors, Surge Protectors, and more**—**derives its contract from this package**.

Any **new functionality or extension** in Electrician **must first be defined here** before being implemented elsewhere.

---

## 📦 Package Overview

This package defines **all component interfaces**, ensuring a **consistent contract** across Electrician.

| File                      | Description                                                                         |
| ------------------------- | ----------------------------------------------------------------------------------- |
| **adapter.go**            | Defines the **Adapter** interface for external data sources.                        |
| **circuitbreaker.go**     | Defines the **CircuitBreaker** interface for failure handling.                      |
| **codec.go**              | Provides interfaces for **Encoders/Decoders**.                                      |
| **common.go**             | Contains **shared types and helper structures**.                                    |
| **conduit.go**            | Defines the **Conduit** interface for multi-stage pipelines.                        |
| **element.go**            | Defines **Element** structs used for message handling.                              |
| **forwardrelay.go**       | Provides the **ForwardRelay** interface for outgoing gRPC messages.                 |
| **generator.go**          | Defines **Generator** components responsible for producing data streams.            |
| **httpclient_adapter.go** | Interface definition for **HTTP-based data ingestion**.                             |
| **logger.go**             | Standardized **logging interfaces** for structured event tracking.                  |
| **meter.go**              | Defines **Meter** interfaces for tracking throughput and performance metrics.       |
| **plug.go**               | Specifies the **Plug** interface for pluggable input sources.                       |
| **receiver.go**           | Defines the **Receiver** interface for incoming message handling.                   |
| **receivingrelay.go**     | Interface for handling **incoming gRPC messages**.                                  |
| **resister.go**           | Defines **Resister** logic for rate-limiting and delayed message processing.        |
| **sensor.go**             | Defines **Sensors** for monitoring and telemetry collection.                        |
| **submitter.go**          | Standardized interface for components that **submit** elements into processing.     |
| **surgeprotector.go**     | Provides the **SurgeProtector** interface for load balancing and overflow handling. |
| **wave.go**               | Specialized **Wave Encoding** interface (for advanced encoding scenarios).          |
| **wire.go**               | Defines the **Wire** interface—the backbone of Electrician.                         |

---

## 🔗 Relationship with Other Packages

The **types** package is **foundational** to Electrician and directly influences every component:

1️⃣ **Internal packages (e.g., `wire`, `circuitbreaker`, `sensor`)** must conform to these interfaces.  
2️⃣ **The builder package (`pkg/builder/`)** wraps and exposes these types via a **user-friendly API**.  
3️⃣ **All components must implement these contracts** to ensure seamless interoperability.

By centralizing all **interfaces and contracts**, Electrician maintains **strict consistency and type safety**.

---

## 🔧 Modifying or Extending a Component

To **add new functionality** in Electrician, follow this structured **workflow**:

### 1️⃣ Modify `types/`

- **Define the new interface method** in the appropriate file (e.g., `wire.go`, `circuitbreaker.go`).
- This ensures **all implementations remain consistent**.

### 2️⃣ Implement in `api.go`

- The `api.go` file inside the respective **internal package** must now implement this method.

### 3️⃣ Add a Functional Option in `options.go`

- Supports **composable, declarative-style configuration**.

### 4️⃣ Extend `notify.go` for event logging (if applicable)

- If your change introduces **new events**, add corresponding **logging and telemetry hooks**.

### 5️⃣ Unit Testing (`package_test.go`)

- **Write tests** to verify that the new functionality works as expected.

By following these steps, Electrician maintains **stability, compatibility, and strict type safety**.

---

## ⚡ Standard Library First

Like most of Electrician, the **Types package is built entirely on Go’s standard library**.  
This ensures:

✅ **Maximum compatibility** – No unnecessary third-party dependencies.  
✅ **Minimal attack surface** – Secure and easy to audit.  
✅ **High performance** – Optimized for **low-latency, high-throughput pipelines**.

Electrician adheres to a **strict standard-library-first** philosophy, ensuring long-term maintainability.

---

## 📖 Further Reading

- **[Root README](../../../README.md)** – Electrician’s overall architecture and principles.
- **[Internal README](../README.MD)** – How `internal/` packages interact with `types/`.
- **[Examples Directory](../../../example/)** – Demonstrates how these interfaces power Electrician's features.

---

## 📝 License

The **Types package** is part of Electrician and is released under the [Apache 2.0 License](../../../LICENSE).  
You’re free to use, modify, and distribute it within these terms.

---

## ⚡ Happy wiring! 🚀

If you have any questions or need support, feel free to **open a GitHub issue**.
