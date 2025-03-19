# 🏗️ Builder Package

The **builder** package provides the **public API** for constructing and configuring Electrician’s core components. It acts as the **entry point** for users to create, customize, and assemble pipelines using functional options.

Instead of directly interacting with **internal** packages, users interface with Electrician **exclusively** through the builder package, which ensures a **clean, composable, and user-friendly** experience.

---

## 📦 Package Overview

The builder package exposes **high-level factories** for each Electrician component. These functions provide a structured way to create and configure objects without requiring users to manage **internal details**.

| File                  | Description                                                                            |
| --------------------- | -------------------------------------------------------------------------------------- |
| **adapter.go**        | Creates external data source adapters (e.g., HTTP, OAuth2).                            |
| **builder.go**        | Houses core logic for constructing and initializing components.                        |
| **circuitbreaker.go** | Provides an interface for failure detection and recovery mechanisms.                   |
| **codec.go**          | Enables encoding and decoding support for data serialization.                          |
| **conduit.go**        | Constructs **multi-stage processing units** by linking multiple Wires together.        |
| **forwardrelay.go**   | Builds components for **sending** data via gRPC (outgoing messages).                   |
| **generator.go**      | Configures autonomous **data generators** that feed Wires.                             |
| **logger.go**         | Exposes logging functionalities for debugging and monitoring.                          |
| **meter.go**          | Provides **performance metrics**, throughput tracking, and concurrency monitoring.     |
| **plug.go**           | Constructs **pluggable data sources** (like adapters or functions) for pipelines.      |
| **receivingrelay.go** | Builds components for **receiving** external data via gRPC.                            |
| **resister.go**       | Implements queue-based **rate limiting** and message resisters.                        |
| **sensor.go**         | Creates sensors that collect **telemetry and real-time monitoring** data.              |
| **surgeprotector.go** | Constructs **load-balancing and overload prevention** components.                      |
| **utils.go**          | Provides helper functions for Electrician’s internal operations.                       |
| **wire.go**           | Constructs the **core data processing units** (Wires) and enables their configuration. |

---

## ⚙️ Functional Options & Configuration

Electrician’s **builder API** follows the **functional options pattern**, allowing users to configure components in a declarative, readable way. Instead of requiring complex constructor parameters, the builder provides **composable options**.

For example, when constructing a **Wire**, users can configure it via options:

```go
wire := builder.NewWire[string](
    ctx,
    builder.WireWithCircuitBreaker(circuitBreaker),
    builder.WireWithGenerator(generator),
    builder.WireWithTransformer(transformFunc),
    builder.WireWithLogger(logger),
)
```

## 🚀 Why Use the Builder Package?

✅ **Encapsulation** – Shields users from internal implementation details.  
✅ **Immutability** – Encourages pre-configured, deterministic pipeline construction.  
✅ **Declarative Composition** – Options provide composable, functional-style configuration.  
✅ **Extensibility** – New features can be added without breaking existing APIs.  
✅ **Simplicity** – Eliminates complex struct initialization, reducing user error.

---

## 🏗️ How It Works

Each function in `builder/` serves as a **factory** that wraps the corresponding **internal** component.

For example, creating a **Wire** follows this pattern:

### 1️⃣ Instantiate the Component (Internal Call)

- `NewWire[T]` calls the **internal** `wire.NewWire[T]` function.

### 2️⃣ Apply Functional Options

- Each `WireWithOption()` function **wraps internal API calls**, allowing seamless configuration.

### 3️⃣ Return a Fully Configured Instance

- The user gets a **ready-to-use Wire** without modifying internal code.

### 🔹 Example: Wire Construction

The following builder function:

```go
func NewWire[T any](ctx context.Context, options ...types.Option[types.Wire[T]]) types.Wire[T] {
    return wire.NewWire[T](ctx, options...)
}
```

- Calls `wire.NewWire[T]` from the **internal** package.
- Accepts **functional options** that modify its behavior.
- Returns a fully configured **Wire** object.

Similarly, each Electrician component—**Circuit Breakers, Surge Protectors, Sensors, etc.**—follows the same structured API.

---

## 🔧 Extending the Builder Package

To **add new functionality**, follow this structured **workflow**:

### 1️⃣ Modify `types/`

- If a new method is needed, **update the interface** in `types/`.
- This ensures **consistent contracts** across components.

### 2️⃣ Implement in `internal/`

- Add the actual method in the respective **internal package (`api.go`)**.

### 3️⃣ Create a Builder Function

- Define a `NewComponent()` function in `builder/`.

### 4️⃣ Expose Functional Options

- Add corresponding `ComponentWithOption()` functions for composability.

By following these steps, Electrician maintains **uniformity** and **backward compatibility** across all components.

---

## 📖 Further Reading

- **[Root README](../../README.md)** – Electrician’s overall architecture and principles.
- **[Internal README](../internal/README.md)** – In-depth explanation of how the `internal/` directory works.
- **[Examples Directory](../../example/)** – Real-world use cases demonstrating the builder API in action.

---

## 📝 License

The **builder package** is part of Electrician and is released under the [Apache 2.0 License](../../LICENSE).  
You’re free to use, modify, and distribute it within these terms.

---

**Happy wiring! ⚙️🚀** If you have questions or need support, feel free to open a GitHub issue.
