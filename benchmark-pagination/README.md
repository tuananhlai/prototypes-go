Basically, execution time for limit-offset queries increase linearly, while cursor-based queries take similar time throughout.

```sh
➜  benchmark-pagination git:(main) ✗ go test -bench .
goos: darwin
goarch: arm64
pkg: github.com/tuananhlai/prototypes/benchmark-pagination
cpu: Apple M1
BenchmarkLimitOffset0-8           5030	   281610 ns/op
BenchmarkLimitOffset10000-8        196	  6126887 ns/op
BenchmarkLimitOffset90000-8        210	  5872436 ns/op
BenchmarkCursor0-8                3972	   297919 ns/op
BenchmarkCursor10000-8            4266	   294308 ns/op
BenchmarkCursor90000-8            4292	   279203 ns/op
PASS
ok  	github.com/tuananhlai/prototypes/benchmark-pagination21.9
```
