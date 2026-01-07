Go Tool Cheatsheet
==================

Quick reference for `go tool` and the most commonly used subcommands. Run
`go tool` with no args to list tools available in your Go version. On this
machine the list currently includes: `asm`, `cgo`, `compile`, `cover`,
`link`, `preprofile`, `vet`.

Basics
------
- List tools: `go tool`
- Run a tool: `go tool <name> [args]`
- Tool help: `go tool <name> -h`
- Use package import path with tools that accept it, or file paths for
  low-level tools.

Common profiling workflows
--------------------------
- CPU profile (pprof): `go test -bench . -cpuprofile cpu.pprof`
- Heap profile (pprof): `go test -bench . -memprofile mem.pprof`
- View profile: `go tool pprof -http=:8080 ./pkg.test cpu.pprof`
- Text summary: `go tool pprof -top ./pkg.test cpu.pprof`
- Trace: `go test -run . -trace trace.out` then `go tool trace trace.out`

Coverage workflows
------------------
- Create cover profile: `go test -coverprofile=cover.out ./...`
- HTML report: `go tool cover -html=cover.out`
- Function summary: `go tool cover -func=cover.out`
- Merge/inspect new format: `go tool covdata textfmt -i=dir -o=cover.out`

Core tools (high frequency)
---------------------------
- `pprof`: CPU/heap/alloc/block/mutex profiling and visualization.
  - `go tool pprof -http=:8080 <binary> <profile>`
  - `go tool pprof -top <binary> <profile>`
- `trace`: Go execution trace viewer.
  - `go tool trace trace.out`
- `cover`: Coverage reports from `go test -coverprofile`.
  - `go tool cover -html=cover.out`
- `test2json`: Convert `go test` output to JSON.
  - `go test -json ./... | go tool test2json -t`

Binary inspection and debugging
-------------------------------
- `addr2line`: Map program counters to file:line.
  - `go tool addr2line -e <bin> <pc>`
- `nm`: List symbols in object files or binaries.
  - `go tool nm -size -sort address <bin>`
- `objdump`: Disassemble and show Go assembly.
  - `go tool objdump -s main.main <bin>`
- `buildid`: Print build ID of a binary.
  - `go tool buildid <bin>`

Assembly and compiler internals
-------------------------------
- `asm`: Go assembler for `.s` files.
  - `go tool asm -o file.o file.s`
- `compile`: Go compiler for a single package.
  - `go tool compile -o pkg.a -p pkg ./file.go`
- `link`: Go linker to produce binaries.
  - `go tool link -o app pkg.a`
- `cgo`: Process cgo source files.
  - `go tool cgo file.go`

Packaging and archives
----------------------
- `pack`: Create and list Go archives (`.a`).
  - `go tool pack r lib.a file.o`
  - `go tool pack t lib.a`

Docs and source maintenance
---------------------------
- `doc`: Display Go doc for packages, types, and symbols.
  - `go tool doc fmt.Fprintf`
- `fix`: Apply fixes for newer Go versions.
  - `go tool fix ./...`
- `vet`: Run Go vet (static checks).
  - `go tool vet ./...`

Distribution and SDK internals (less common)
--------------------------------------------
- `dist`: Build the Go toolchain from source (Go repo).
- `covdata`: Inspect and merge coverage data directories.
  - `go tool covdata textfmt -i=dir -o=cover.out`
- `preprofile`: Experimental tool (may appear in some Go builds).

Full tool catalog (varies by Go version)
----------------------------------------
- Profiling/trace: `pprof`, `trace`, `test2json`
- Binary inspection: `addr2line`, `nm`, `objdump`, `buildid`
- Archives/build: `pack`, `fix`, `doc`, `covdata`, `dist`
- Core toolchain: `asm`, `cgo`, `compile`, `link`, `vet`, `cover`

Tips
----
- Use `go tool` tools directly for lower-level workflows; prefer `go test`,
  `go vet`, and `go doc` for high-level use.
- For profiling, supply the test binary (`./pkg.test`) or built binary so
  `pprof` can resolve symbols.
- Tool availability depends on your Go version; check `go tool` output.
