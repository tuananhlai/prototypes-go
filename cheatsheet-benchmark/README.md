# Go Benchmarking – `go test` Command Cheatsheet

Below is a **practical, copy-paste-ready cheatsheet** of `go test` commands for **benchmarking** in Go.

---

## 1️⃣ Run all benchmarks (skip tests)

```bash
go test -run '^$' -bench .
```

- `-run '^$'` → run no tests
- `-bench .` → run all benchmarks in the package

This is the **default starting command**.

---

## 2️⃣ Run all benchmarks with allocation stats

```bash
go test -run '^$' -bench . -benchmem
```

Adds:

- `B/op`
- `allocs/op`

Almost always what you want.

---

## 3️⃣ Run a single benchmark

```bash
go test -run '^$' -bench BenchmarkOnDuplicateKeyUpdate
```

Benchmarks are matched by **regex**.

---

## 4️⃣ Run multiple specific benchmarks

```bash
go test -run '^$' -bench 'Benchmark(OnDuplicateKeyUpdate|ReplaceInto)'
```

Useful for focused comparisons.

---

## 5️⃣ Increase benchmark duration (reduce noise)

```bash
go test -run '^$' -bench . -benchtime=3s
```

Default is ~1s per benchmark.  
Longer time = more stable results.

---

## 6️⃣ Fixed iteration count (not time-based)

```bash
go test -run '^$' -bench . -benchtime=10000x
```

Useful when:

- using `StopTimer`
- benchmarking extremely fast code
- avoiding “benchmark never finishes” scenarios

---

## 7️⃣ Repeat benchmarks multiple times

```bash
go test -run '^$' -bench . -benchmem -count=5
```

Runs each benchmark 5 times.  
Great for statistical comparison.

---

## 8️⃣ Compare results with `benchstat`

```bash
go test -run '^$' -bench . -benchmem -count=5 > old.txt
go test -run '^$' -bench . -benchmem -count=5 > new.txt
benchstat old.txt new.txt
```

This is the **gold standard** for performance comparison.

---

## 9️⃣ Run benchmarks with different CPU counts

```bash
go test -run '^$' -bench . -cpu=1,2,4,8
```

- Changes `GOMAXPROCS`
- Appends `-1`, `-2`, `-4`, etc. to benchmark names

Great for:

- lock contention
- parallel algorithms
- DB connection pool tuning

---

## 🔟 Parallel benchmarks only

```bash
go test -run '^$' -bench BenchmarkAtomicVsMutex -cpu=8
```

Useful when testing `b.RunParallel`.

---

## 1️⃣1️⃣ Benchmark a specific package

```bash
go test -run '^$' -bench . ./pkg/storage
```

Or recursively:

```bash
go test -run '^$' -bench . ./...
```

⚠️ Recursive benchmarking can be **very slow**.

---

## 1️⃣2️⃣ Save CPU & memory profiles

```bash
go test -run '^$' -bench BenchmarkX \
  -cpuprofile cpu.out \
  -memprofile mem.out
```

View them:

```bash
go tool pprof -http=:0 cpu.out
go tool pprof -http=:0 mem.out
```

---

## 1️⃣3️⃣ JSON output (for tooling / CI)

```bash
go test -run '^$' -bench . -json > bench.json
```

Used for:

- dashboards
- CI regression detection
- custom analysis tools

---

## 1️⃣4️⃣ Disable optimizations (debug only)

```bash
go test -run '^$' -bench . -gcflags=all='-N -l'
```

⚠️ Results will not reflect real performance.  
Use only to understand optimizer behavior.

---

## 1️⃣5️⃣ Verbose benchmark output

```bash
go test -run '^$' -bench . -v
```

Mostly useful if your benchmark logs errors.

---

## 1️⃣6️⃣ Real-world “serious benchmark” command

```bash
go test -run '^$' \
  -bench 'Benchmark(OnDuplicateKeyUpdate|ReplaceInto)' \
  -benchmem \
  -benchtime=3s \
  -count=5 \
  -cpu=1
```

---

## TL;DR — the 5 commands you’ll actually use

```bash
go test -run '^$' -bench . -benchmem
go test -run '^$' -bench BenchmarkX
go test -run '^$' -bench . -benchmem -count=5
go test -run '^$' -bench . -benchtime=3s
benchstat old.txt new.txt
```

---
