# 🛠️ Resister Package – Priority-Based Queueing

The **Resister package** implements a **thread-safe, priority queue** for managing **delayed or rate-limited data flow** in Electrician pipelines. It serves as an **intelligent buffer**, ensuring **high-priority elements are processed first**, while **lower-priority elements decay over time**.

Resisters are particularly useful in **congested systems** where processing capacity is limited, and **queued elements must be dynamically prioritized**.

---

## 📦 Package Overview

| Feature                         | Description                                                               |
| ------------------------------- | ------------------------------------------------------------------------- |
| **Priority-Based Queueing**     | Uses a **priority heap** to process critical elements first.              |
| **Retry & Requeue Support**     | Elements **requeued on failure** retain their priority but can decay.     |
| **Concurrency-Safe**            | Thread-safe operations for concurrent pipeline execution.                 |
| **Backoff & Priority Decay**    | Items that **fail repeatedly** have their priority **gradually reduced**. |
| **Sensor & Logger Integration** | Supports **event monitoring** and **structured logging**.                 |

---

## 📂 Package Structure

Each file follows **Electrician’s structured approach**, ensuring **clear separation of concerns**.

| File                 | Purpose                                                                        |
| -------------------- | ------------------------------------------------------------------------------ |
| **api.go**           | Public API for interacting with **priority queue operations**.                 |
| **internal.go**      | Manages internal logic for **queue indexing, priority decay, and processing**. |
| **notify.go**        | Handles **event logging, telemetry, and sensor notifications**.                |
| **options.go**       | Functional options for configuring Resisters **declaratively**.                |
| **resister.go**      | Core **Type Definition and Constructor**.                                      |
| **resister_test.go** | Unit tests ensuring **correctness and performance**.                           |

---

## 🔧 How Resisters Work

A **Resister** functions as a **priority queue**, ensuring that the **most important elements are processed first**.  
It dynamically **decays priority over time** to prevent starvation of lower-priority items.

### ✅ **Key Mechanisms**

- **Heap-Based Priority Queue** – Ensures elements are processed in the correct order.
- **Requeue with Adjusted Priority** – Failed elements **re-enter the queue** but **lose priority over time**.
- **Decay Mechanism** – Elements that **fail too often** will **gradually drop in priority**.
- **Concurrency-Safe Processing** – Locking mechanisms ensure safe **multi-threaded access**.
- **Queue Monitoring** – Sensors track queue **size, processing rates, and failure rates**.

---

## 🔒 Standard Library First

Like most of Electrician, the **Resister package is built entirely on Go’s standard library**, ensuring:

✅ **Maximum compatibility** – No unnecessary third-party dependencies.  
✅ **Minimal attack surface** – Secure and easy to audit.  
✅ **High performance** – Optimized for **low-latency, high-throughput queueing**.

Electrician follows a **strict standard-library-first** approach, ensuring long-term maintainability.

---

## 🔧 Extending the Resister Package

To **add new functionality** to the Resister package, follow this structured **workflow**:

### 1️⃣ Modify `types/`

- **Define the new method** inside `types/resister.go`.
- This ensures **all implementations remain consistent**.

### 2️⃣ Implement the logic in `api.go`

- The `api.go` file inside the **resister** package must now implement this method.

### 3️⃣ Add a Functional Option in `options.go`

- Supports **declarative-style queue configuration**.

### 4️⃣ Ensure `notify.go` handles event logging (if applicable)

- If your change introduces **new queue-related events**, add corresponding **logging and telemetry hooks**.

### 5️⃣ Unit Testing (`resister_test.go`)

- **Write tests** to verify the **priority queue logic and decay mechanisms**.

By following these steps, Electrician maintains **stability, compatibility, and strict type safety**.

---

## 📖 Further Reading

- **[Root README](../../../README.md)** – Electrician’s overall architecture and principles.
- **[Internal README](../README.MD)** – How `internal/` packages interact with `types/`.
- **[Examples Directory](../../../example/surge_protector_example/rate_limit/)** – Demonstrates how Resisters **prioritize, retry, and decay elements** in real-world use cases.

---

## 📝 License

The **Resister package** is part of Electrician and is released under the [Apache 2.0 License](../../../LICENSE).  
You’re free to use, modify, and distribute it within these terms.

---

## ⚡ Happy wiring! 🚀

If you have any questions or need support, feel free to **open a GitHub issue**.
