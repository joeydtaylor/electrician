# ⚙️ Utils Package

The **utils** package provides **general-purpose utility functions** used throughout Electrician.  
Unlike other internal packages, `utils/` does **not** define Electrician components—  
instead, it centralizes **helper functions** for **functional operations, file handling, hashing, and testing**.

This package exists to **reduce code duplication** and make Electrician **more maintainable**.

---

## 📦 Package Overview

The `utils` package contains **core utilities** that support multiple internal components.

| Feature                | Description                                                         |
| ---------------------- | ------------------------------------------------------------------- |
| **Functional Helpers** | Implements **Map, Filter, and Contains** for slice transformations. |
| **File Handling**      | Provides **Unzip, Untar**, and other archive utilities.             |
| **Hashing & Security** | Implements **SHA-256 hashing** and **unique hash generation**.      |
| **Test Utilities**     | Provides **structured test setup and teardown** mechanisms.         |

---

## 📂 Package Structure

| File                 | Purpose                                                            |
| -------------------- | ------------------------------------------------------------------ |
| **functors.go**      | Implements **functional utilities** like `Map()` and `Filter()`.   |
| **utils.go**         | General **helper functions** that don’t fit into other categories. |
| **utils_test.go**    | Contains **tests** for utility functions.                          |
| **functors_test.go** | Unit tests **specifically for functional utilities**.              |
| **main_test.go**     | Entry point for **test setup and teardown** processes.             |

---

## 🔧 Utilities & Their Usage

### 🔹 **Functional Helpers**

Electrician includes **functional programming utilities**, making it easy to work with slices.  
These utilities allow for **filtering, mapping, and checking membership** in collections.

- **Map()** – Applies a function to each element in a slice.
- **Filter()** – Returns only elements that satisfy a condition.
- **Contains()** – Checks if a slice contains a specific element.

### 🔹 **File Handling Utilities**

Electrician provides **archive utilities** for working with compressed files.  
These are useful for **managing dependencies** in testing and automation.

- **Unzip()** – Extracts `.zip` archives.
- **Untar()** – Extracts `.tar.gz` archives.

### 🔹 **Hashing & Security**

Electrician requires **deterministic hashing** for **tracking components and generating unique identifiers**.

- **GenerateSha256Hash()** – Computes a **SHA-256 hash** from any input.
- **GenerateUniqueHash()** – Generates a **randomized, cryptographically strong identifier**.

These utilities ensure **fast, secure hashing** for **component tracking and deduplication**.

---

## ⚡ Standard Library First

The **Utils package**, like most of Electrician, is **built entirely on Go’s standard library**.  
It **does not** introduce **any external dependencies**—ensuring:

✅ **Maximum compatibility** – No unnecessary libraries.  
✅ **Minimal attack surface** – Secure and easy to audit.  
✅ **High performance** – Optimized for **low-latency, high-throughput operations**.

---

## 🔧 Extending the Utils Package

To **add a new utility function**, follow this structured **workflow**:

1️⃣ **Determine if it belongs in `utils/`**

- Utilities should be **generic** and **widely reusable**.
- If the function is **specific to a single component**, it **belongs in that package instead**.

2️⃣ **Implement in `utils.go` or `functors.go`**

- **General-purpose helpers** go in `utils.go`.
- **Slice and functional utilities** belong in `functors.go`.

3️⃣ **Write Tests in `utils_test.go` or `functors_test.go`**

- Every function **must** have **unit tests** before merging.

By following these **strict contribution guidelines**, Electrician remains **clean, efficient, and maintainable**.

---

## 📖 Further Reading

- **[Root README](../../../README.md)** – Electrician’s overall architecture and principles.
- **[Internal README](../README.MD)** – How `internal/` packages interact with `types/`.
- **[Examples Directory](../../../example/)** – Real-world use cases demonstrating Electrician’s functionality.

---

## 📝 License

The **Utils package** is part of Electrician and is released under the [Apache 2.0 License](../../../LICENSE).  
You’re free to use, modify, and distribute it within these terms.

---

## ⚡ Happy wiring! 🚀

If you have any questions or need support, feel free to **open a GitHub issue**.
