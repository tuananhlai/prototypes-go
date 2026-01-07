#!/usr/bin/env sh
set -eu

name="${1:-}"
if [ -z "$name" ]; then
  echo "Usage: task g -- <folder>"
  exit 1
fi

dir="[todo] $name"
mkdir -p "$dir"
printf '%s\n' "module github.com/tuananhlai/prototypes/$name" "" "go 1.25.5" > "$dir/go.mod"
printf '%s\n' \
  "package main" \
  "" \
  "func main() {" \
  "" \
  "}" > "$dir/main.go"
