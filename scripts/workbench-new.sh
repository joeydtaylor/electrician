#!/usr/bin/env bash
set -euo pipefail

if [[ ${1:-} == "" ]]; then
  echo "usage: scripts/workbench-new.sh <name>" >&2
  exit 1
fi

name="$1"
root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
work_dir="$root_dir/workbench/$name"

if [[ -e "$work_dir" ]]; then
  echo "error: $work_dir already exists" >&2
  exit 1
fi

mkdir -p "$work_dir"

cat <<'GO' > "$work_dir/main.go"
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

(
  cd "$work_dir"
  go mod init "example.com/$name" >/dev/null
  go mod edit -require github.com/joeydtaylor/electrician@v0.0.0
  go mod edit -replace github.com/joeydtaylor/electrician=../../
)

echo "created $work_dir"
echo "run: cd workbench/$name && go run ."
