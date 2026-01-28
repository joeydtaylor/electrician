#!/usr/bin/env bash
set -euo pipefail

if [[ ${1:-} == "" ]]; then
  echo "usage: scripts/workbench-build.sh <name>" >&2
  exit 1
fi

name="$1"
root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
work_dir="$root_dir/workbench/$name"
output_dir="$root_dir/out"

if [[ ! -d "$work_dir" ]]; then
  echo "error: $work_dir does not exist" >&2
  exit 1
fi

mkdir -p "$output_dir"
(
  cd "$work_dir"
  go build -o "$output_dir/$name"
)

echo "built $output_dir/$name"
