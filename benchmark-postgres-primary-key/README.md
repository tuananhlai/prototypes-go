- Generating UUID using application codes instead of database is faster.
- Inserts to a table with BIGSERIAL primary is noticably faster than one with UUID primary key. At minimum it is ~30% faster.
- Inserts to a table with UUIDv7 primary is noticably faster than one with UUIDv4 key.
- Size of indexes for UUID table is noticably (50%) larger than BIGSERIAL table.

```sh
➜  benchmark-postgres-primary-key git:(main) ✗ go test -bench .
goos: darwin
goarch: arm64
pkg: github.com/tuananhlai/prototypes/uuid-primary-key-benchmark
cpu: Apple M1 Max
BenchmarkInsertIntPk-10                        2         782317375 ns/op
BenchmarkInsertUUIDv4PkDBGen-10                1        2600753375 ns/op
BenchmarkInsertUUIDv4PkAppGen-10               1        2043274875 ns/op
BenchmarkInsertUUIDv7PkAppGen-10               2         994667729 ns/op
PASS
ok      github.com/tuananhlai/prototypes/uuid-primary-key-benchmark     8.822s
```

```sh
postgres@localhost:postgres> SELECT
     pg_size_pretty(pg_relation_size('bench_uuid_pk_pkey'));
+----------------+
| pg_size_pretty |
|----------------|
| 30 MB          |
+----------------+
SELECT 1
Time: 0.018s
postgres@localhost:postgres> SELECT
     pg_size_pretty(pg_relation_size('bench_int_pk_pkey'));
+----------------+
| pg_size_pretty |
|----------------|
| 21 MB          |
+----------------+
SELECT 1
Time: 0.008s
```