<https://go.dev/wiki/LearnTesting>
<https://go.dev/talks/2014/testing.slide>

```txt
--- Common patterns cheat notes (in comments) ---

- Unit Testing basics:
    // Function under test: func MyFunc(x int) int
    func TestMyFunc(t *testing.T) {
        got := MyFunc(2)
        want := 4
        if got != want {
            t.Errorf("MyFunc(2) = %d; want %d", got, want)
        }
    }
- Table-driven tests:
    func TestSomething(t *testing.T) {
        tests := []struct {
            name string
            input int
            want  int
        }{
            {"double 2", 2, 4},
            {"double 3", 3, 6},
        }
        for _, tc := range tests {
            t.Run(tc.name, func(t *testing.T) {
                got := MyFunc(tc.input)
                if got != tc.want {
                    t.Errorf("MyFunc(%d) = %d; want %d", tc.input, got, tc.want)
                }
            })
        }
    }
- Setup/teardown with t.Cleanup:
    func TestSomething(t *testing.T) {
        t.Cleanup(func() {
            // code to run after test
        })
        // ... test logic ...
    }

- Skipping:
    t.Skip("reason")
- Parallel tests:
    t.Parallel() // careful with shared global state
- Short mode:
    if testing.Short() { t.Skip("skipping in -short") }
- Temp dirs/files:
    dir := t.TempDir()
- Environment:
    t.Setenv("KEY", "VALUE") // Go 1.17+
- Build tags:
    //go:build integration
    Run with: go test -tags=integration ./...
- Testdata convention:
    Put fixtures under: ./testdata (Go tooling ignores it for packages)
- Race detector:
    go test -race ./...
- Coverage:
    go test -cover ./...
- External test package:
    package yourpkg_test // forces testing public API only
- Generate coverage report:
    go test -coverprofile=coverage.out ./...
- Display coverage report as HTML page:
    go tool cover -html=coverage.out

- Run specific tests:
    go test -run '^TestFuncName$'
    go test -run 'SubtestName'                # run subtest by name
    go test -run 'TestName1|TestName2'        # regex for multiple tests
    # For a package:
    go test -run '^TestXxx$' ./path/to/pkg
    # Run with verbose output:
    go test -v -run '^TestFuncName$'

- Benchmarking:
    func BenchmarkXxx(b *testing.B) { ... }
    Example:
        func BenchmarkAdd(b *testing.B) {
            for b.Loop() {
                Add(1, 2)
            }
        }
- Run all benchmarks:
    go test -bench=.
- Run specific benchmark:
    go test -bench=BenchmarkAdd
- Set number of benchmark runs:
    go test -bench=BenchmarkAdd -benchtime=10s
- Show allocations:
    go test -bench=BenchmarkAdd -benchmem


- Fuzzing (Go 1.18+):
    func FuzzXxx(f *testing.F) { ... }
    Example:
        func FuzzMyFunc(f *testing.F) {
            f.Add("example input")
            f.Fuzz(func(t *testing.T, arg string) {
                // test logic
            })
        }
- Run specific fuzz target:
    go test -fuzz=FuzzMyFunc
- Run all fuzz targets:
    go test -fuzz=.
- Set fuzz test duration:
    go test -fuzz=FuzzMyFunc -fuzztime=10s
- Use fuzz corpus files:
    Place corpus inputs under: testdata/fuzz/FuzzMyFunc/
    Each file is a unique input
- View fuzz cache/corpus:
    go test -fuzz=FuzzMyFunc -fuzzcache=<dir>
- Stop after N fuzzing failures:
    go test -fuzz=FuzzMyFunc -failfast

```
