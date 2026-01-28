# Workbench (Instant Codegen Playground)

This repo is designed so you (or a friend with Codex) can generate code quickly
without touching production packages. Use the **workbench** folder for local
experiments, one-off binaries, and "just make it work" prototypes.

> `workbench/` is git-ignored. Nothing here is committed.

---

## 60-second setup (copy/paste)

Fastest path (script):

```bash
./scripts/workbench-new.sh hello
cd workbench/hello
go run .
```

Manual path (copy/paste):

```bash
mkdir -p workbench/hello
cd workbench/hello

go mod init example.com/hello

go mod edit -require github.com/joeydtaylor/electrician@v0.0.0

go mod edit -replace github.com/joeydtaylor/electrician=../../

cat <<'GO' > main.go
package main

import (
  "context"
  "fmt"
  "time"

  "github.com/joeydtaylor/electrician/pkg/builder"
)

type Item struct {
  Message string `json:"message"`
}

func main() {
  ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
  defer cancel()

  w := builder.NewWire[Item](ctx)
  _ = w.Start(ctx)

  _ = w.Submit(ctx, Item{Message: "hello"})

  select {
  case v := <-w.GetOutputChannel():
    fmt.Println("output:", v.Message)
  case <-ctx.Done():
  }
}
GO

# run it

go run .
```

## Build a portable binary

```bash
./scripts/workbench-build.sh hello
./out/hello
```

## Optional: build a C ABI (advanced)

If you need a shared library ABI, Go supports `c-shared` and `c-archive`.
This requires `package main` and exported functions (`//export`).

Example:

```go
//export Add
func Add(a, b int32) int32 { return a + b }
```

Build (Linux example):

```bash
go build -buildmode=c-shared -o ../../out/libhello.so
```

## Codex prompt template

When you ask Codex to generate code, use a prompt like:

> "Create a new example under `workbench/foo` that uses the builder to receive JSON, encrypt with AES-GCM, and send to Relay. Keep it runnable with `go run .`."

Then run:

```bash
go run ./workbench/foo
```

## Clean up

```bash
rm -rf workbench/hello
```
