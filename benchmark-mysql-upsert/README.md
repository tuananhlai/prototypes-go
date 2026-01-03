`REPLACE INTO` does a delete + insert, while `ON DUPLICATE KEY UPDATE` does not.

When the target table only has a primary key column, the two commands perform similarly.

```sh
➜  benchmark-mysql-upsert git:(main) ✗ go test -bench .
goos: darwin
goarch: arm64
pkg: github.com/tuananhlai/prototypes/benchmark-mysql-upsert
cpu: Apple M1
BenchmarkOnDuplicateKeyUpdate-8   	   9177	    126861 ns/op	     192 B/op	       6 allocs/op
BenchmarkReplaceInto-8            	   9478	    126071 ns/op	     192 B/op	       6 allocs/op
PASS
ok  	github.com/tuananhlai/prototypes/benchmark-mysql-upsert	3.023s
```

However, by adding a new index on the target table, the performance diverges noticably.

```sh
➜  benchmark-mysql-upsert git:(main) ✗ go test -bench .
goos: darwin
goarch: arm64
pkg: github.com/tuananhlai/prototypes/benchmark-mysql-upsert
cpu: Apple M1
BenchmarkOnDuplicateKeyUpdate-8   	   9123	    130357 ns/op	     192 B/op	       6 allocs/op
BenchmarkReplaceInto-8            	  10000	    209009 ns/op	     192 B/op	       6 allocs/op
PASS
ok  	github.com/tuananhlai/prototypes/benchmark-mysql-upsert	3.579s
```
