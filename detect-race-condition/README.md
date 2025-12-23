# Detecting data races

## Build the program with data race detection

```
➜  detect-race-condition git:(main) ✗ go build -race . && ./detect-race-condition
==================
WARNING: DATA RACE
Read at 0x00c000010128 by goroutine 7:
  github.com/tuananhlai/prototypes/detect-race-condition/counter.CountToFour.func1()
      /Users/anhlt/Coding/other/prototypes/detect-race-condition/counter/count4.go:12 +0x38
  sync.(*WaitGroup).Go.func1()
      /Users/anhlt/.asdf/installs/golang/1.25.5/go/src/sync/waitgroup.go:239 +0x54

Previous write at 0x00c000010128 by goroutine 8:
  github.com/tuananhlai/prototypes/detect-race-condition/counter.CountToFour.func1()
      /Users/anhlt/Coding/other/prototypes/detect-race-condition/counter/count4.go:12 +0x48
  sync.(*WaitGroup).Go.func1()
      /Users/anhlt/.asdf/installs/golang/1.25.5/go/src/sync/waitgroup.go:239 +0x54

Goroutine 7 (running) created at:
  sync.(*WaitGroup).Go()
      /Users/anhlt/.asdf/installs/golang/1.25.5/go/src/sync/waitgroup.go:237 +0x78
  github.com/tuananhlai/prototypes/detect-race-condition/counter.CountToFour()
      /Users/anhlt/Coding/other/prototypes/detect-race-condition/counter/count4.go:10 +0x70
  main.main()
      /Users/anhlt/Coding/other/prototypes/detect-race-condition/main.go:10 +0x20

Goroutine 8 (finished) created at:
  sync.(*WaitGroup).Go()
      /Users/anhlt/.asdf/installs/golang/1.25.5/go/src/sync/waitgroup.go:237 +0x78
  github.com/tuananhlai/prototypes/detect-race-condition/counter.CountToFour()
      /Users/anhlt/Coding/other/prototypes/detect-race-condition/counter/count4.go:10 +0x70
  main.main()
      /Users/anhlt/Coding/other/prototypes/detect-race-condition/main.go:10 +0x20
==================
count=4
Found 1 data race(s)
```

## Run unit tests with data race detection

```
➜  detect-race-condition git:(main) ✗ go test -race ./counter
==================
WARNING: DATA RACE
Read at 0x00c00009a188 by goroutine 9:
  github.com/tuananhlai/prototypes/detect-race-condition/counter.CountToFour.func1()
      /Users/anhlt/Coding/other/prototypes/detect-race-condition/counter/count4.go:12 +0x38
  sync.(*WaitGroup).Go.func1()
      /Users/anhlt/.asdf/installs/golang/1.25.5/go/src/sync/waitgroup.go:239 +0x54

Previous write at 0x00c00009a188 by goroutine 8:
  github.com/tuananhlai/prototypes/detect-race-condition/counter.CountToFour.func1()
      /Users/anhlt/Coding/other/prototypes/detect-race-condition/counter/count4.go:12 +0x48
  sync.(*WaitGroup).Go.func1()
      /Users/anhlt/.asdf/installs/golang/1.25.5/go/src/sync/waitgroup.go:239 +0x54

Goroutine 9 (running) created at:
  sync.(*WaitGroup).Go()
      /Users/anhlt/.asdf/installs/golang/1.25.5/go/src/sync/waitgroup.go:237 +0x78
  github.com/tuananhlai/prototypes/detect-race-condition/counter.CountToFour()
      /Users/anhlt/Coding/other/prototypes/detect-race-condition/counter/count4.go:10 +0x70
  github.com/tuananhlai/prototypes/detect-race-condition/counter_test.TestCountToFour()
      /Users/anhlt/Coding/other/prototypes/detect-race-condition/counter/count4_test.go:10 +0x24
  testing.tRunner()
      /Users/anhlt/.asdf/installs/golang/1.25.5/go/src/testing/testing.go:1934 +0x164
  testing.(*T).Run.gowrap1()
      /Users/anhlt/.asdf/installs/golang/1.25.5/go/src/testing/testing.go:1997 +0x3c

Goroutine 8 (finished) created at:
  sync.(*WaitGroup).Go()
      /Users/anhlt/.asdf/installs/golang/1.25.5/go/src/sync/waitgroup.go:237 +0x78
  github.com/tuananhlai/prototypes/detect-race-condition/counter.CountToFour()
      /Users/anhlt/Coding/other/prototypes/detect-race-condition/counter/count4.go:10 +0x70
  github.com/tuananhlai/prototypes/detect-race-condition/counter_test.TestCountToFour()
      /Users/anhlt/Coding/other/prototypes/detect-race-condition/counter/count4_test.go:10 +0x24
  testing.tRunner()
      /Users/anhlt/.asdf/installs/golang/1.25.5/go/src/testing/testing.go:1934 +0x164
  testing.(*T).Run.gowrap1()
      /Users/anhlt/.asdf/installs/golang/1.25.5/go/src/testing/testing.go:1997 +0x3c
==================
--- FAIL: TestCountToFour (0.00s)
    testing.go:1617: race detected during execution of test
FAIL
FAIL	github.com/tuananhlai/prototypes/detect-race-condition/counter	0.199s
FAIL
```
