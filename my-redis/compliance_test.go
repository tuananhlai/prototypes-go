package main

import (
	"bytes"
	"context"
	"net"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

func TestCompliance(t *testing.T) {
	suite.Run(t, new(ComplianceTestSuite))
}

type ComplianceTestSuite struct {
	suite.Suite
	testServerListener net.Listener
}

func (s *ComplianceTestSuite) SetupTest() {
	listener, err := net.Listen("tcp", "127.0.0.1:6379")
	if err != nil {
		s.T().Fatalf("failed to listen: %v", err)
	}

	s.testServerListener = listener

	go run(listener)
}

func (s *ComplianceTestSuite) TearDownTest() {
	s.testServerListener.Close()
}

func (s *ComplianceTestSuite) TestPing() {
	cmd := exec.Command("redis-cli", "PING")
	out, err := cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("PONG\n", string(out))
}

func (s *ComplianceTestSuite) TestMultiplePings() {
	cmd := exec.Command("redis-cli")
	cmd.Stdin = bytes.NewBufferString("PING\nPING\n")

	out, err := cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("PONG\nPONG\n", string(out))
}

func (s *ComplianceTestSuite) TestMultipleClients() {
	numConcurrentClients := 2

	results := make(chan struct {
		out string
		err error
	}, numConcurrentClients)

	var wg sync.WaitGroup
	for range numConcurrentClients {
		wg.Go(func() {
			cmd := exec.Command("redis-cli", "PING")
			out, err := cmd.CombinedOutput()
			results <- struct {
				out string
				err error
			}{string(out), err}
		})
	}

	wg.Wait()
	close(results)

	for result := range results {
		s.Require().NoError(result.err)
		s.Equal("PONG\n", result.out)
	}
}

func (s *ComplianceTestSuite) TestEcho() {
	cmd := exec.Command("redis-cli", "ECHO", "hello")
	out, err := cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("hello\n", string(out))
}

func (s *ComplianceTestSuite) TestSetGet() {
	cmd := exec.Command("redis-cli", "SET", "foo", "bar")
	out, err := cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("OK\n", string(out))

	// Get existing key
	cmd = exec.Command("redis-cli", "GET", "foo")
	out, err = cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("bar\n", string(out))

	// Get non-existent key
	cmd = exec.Command("redis-cli", "GET", "baz")
	out, err = cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("\n", string(out))
}

func (s *ComplianceTestSuite) TestRPushCreatesNewList() {
	cmd := exec.Command("redis-cli", "RPUSH", "list_key", "element")
	out, err := cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("1\n", string(out))
}

func (s *ComplianceTestSuite) TestRPushAppendsToExistingList() {
	cmd := exec.Command("redis-cli", "RPUSH", "list_key", "element1")
	out, err := cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("1\n", string(out))

	cmd = exec.Command("redis-cli", "RPUSH", "list_key", "element2")
	out, err = cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("2\n", string(out))
}

func (s *ComplianceTestSuite) TestRPushMultipleElements() {
	cmd := exec.Command("redis-cli", "RPUSH", "list_key", "element1", "element2", "element3")
	out, err := cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("3\n", string(out))

	cmd = exec.Command("redis-cli", "RPUSH", "list_key", "element4", "element5")
	out, err = cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("5\n", string(out))
}

func (s *ComplianceTestSuite) TestLPushCreatesNewList() {
	cmd := exec.Command("redis-cli", "LPUSH", "list_key", "c")
	out, err := cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("1\n", string(out))
}

func (s *ComplianceTestSuite) TestLPushPrependsToExistingList() {
	cmd := exec.Command("redis-cli", "LPUSH", "list_key", "c")
	out, err := cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("1\n", string(out))

	// A second LPUSH prepends to the existing list and returns the new length.
	cmd = exec.Command("redis-cli", "LPUSH", "list_key", "b")
	out, err = cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("2\n", string(out))

	cmd = exec.Command("redis-cli", "LRANGE", "list_key", "0", "-1")
	out, err = cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("b\nc\n", string(out))
}

func (s *ComplianceTestSuite) TestLPushMultipleElementsAreReversed() {
	cmd := exec.Command("redis-cli", "LPUSH", "list_key", "a", "b", "c")
	out, err := cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("3\n", string(out))

	// Elements pushed left-to-right end up in reverse order in the list.
	cmd = exec.Command("redis-cli", "LRANGE", "list_key", "0", "-1")
	out, err = cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("c\nb\na\n", string(out))
}

func (s *ComplianceTestSuite) TestLPushThenLPushPreservesOrder() {
	cmd := exec.Command("redis-cli", "LPUSH", "list_key", "c")
	out, err := cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("1\n", string(out))

	cmd = exec.Command("redis-cli", "LPUSH", "list_key", "b", "a")
	out, err = cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("3\n", string(out))

	// First LPUSH "c" → [c]; then LPUSH "b" "a" prepends b then a → [a, b, c].
	cmd = exec.Command("redis-cli", "LRANGE", "list_key", "0", "-1")
	out, err = cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("a\nb\nc\n", string(out))
}

func (s *ComplianceTestSuite) TestLRange() {
	cmd := exec.Command("redis-cli", "RPUSH", "list_key", "a", "b", "c", "d", "e")
	out, err := cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("5\n", string(out))

	// First 2 elements.
	cmd = exec.Command("redis-cli", "LRANGE", "list_key", "0", "1")
	out, err = cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("a\nb\n", string(out))

	// Elements from index 2 to 4 (stop is inclusive).
	cmd = exec.Command("redis-cli", "LRANGE", "list_key", "2", "4")
	out, err = cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("c\nd\ne\n", string(out))
}

func (s *ComplianceTestSuite) TestLRangeNonExistentList() {
	cmd := exec.Command("redis-cli", "LRANGE", "missing_key", "0", "1")
	out, err := cmd.CombinedOutput()
	s.Require().NoError(err)
	// redis-cli renders an empty array (*0\r\n) as a single newline.
	s.Equal("\n", string(out))
}

func (s *ComplianceTestSuite) TestLRangeStartBeyondEnd() {
	cmd := exec.Command("redis-cli", "RPUSH", "list_key", "a", "b", "c")
	out, err := cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("3\n", string(out))

	// Start index >= length yields an empty array.
	cmd = exec.Command("redis-cli", "LRANGE", "list_key", "5", "10")
	out, err = cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("\n", string(out))
}

func (s *ComplianceTestSuite) TestLRangeStopBeyondEnd() {
	cmd := exec.Command("redis-cli", "RPUSH", "list_key", "a", "b", "c", "d", "e")
	out, err := cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("5\n", string(out))

	// Stop index >= length is clamped to the last element.
	cmd = exec.Command("redis-cli", "LRANGE", "list_key", "0", "10")
	out, err = cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("a\nb\nc\nd\ne\n", string(out))
}

func (s *ComplianceTestSuite) TestLRangeStartGreaterThanStop() {
	cmd := exec.Command("redis-cli", "RPUSH", "list_key", "a", "b", "c")
	out, err := cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("3\n", string(out))

	// Start index > stop index yields an empty array.
	cmd = exec.Command("redis-cli", "LRANGE", "list_key", "2", "1")
	out, err = cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("\n", string(out))
}

func (s *ComplianceTestSuite) TestLRangeNegativeIndexes() {
	cmd := exec.Command("redis-cli", "RPUSH", "list_key", "a", "b", "c", "d", "e")
	out, err := cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("5\n", string(out))

	// Last 2 elements via negative start and stop.
	cmd = exec.Command("redis-cli", "LRANGE", "list_key", "-2", "-1")
	out, err = cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("d\ne\n", string(out))
}

func (s *ComplianceTestSuite) TestLRangeMixedIndexes() {
	cmd := exec.Command("redis-cli", "RPUSH", "list_key", "a", "b", "c", "d", "e")
	out, err := cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("5\n", string(out))

	// Positive start, negative stop: all items except the last 2.
	cmd = exec.Command("redis-cli", "LRANGE", "list_key", "0", "-3")
	out, err = cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("a\nb\nc\n", string(out))

	// Positive start, -1 stop: from index 2 to the end.
	cmd = exec.Command("redis-cli", "LRANGE", "list_key", "2", "-1")
	out, err = cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("c\nd\ne\n", string(out))
}

func (s *ComplianceTestSuite) TestLRangeNegativeStartOutOfRange() {
	cmd := exec.Command("redis-cli", "RPUSH", "list_key", "a", "b", "c", "d", "e")
	out, err := cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("5\n", string(out))

	// A negative start beyond the list length is treated as 0.
	cmd = exec.Command("redis-cli", "LRANGE", "list_key", "-6", "-1")
	out, err = cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("a\nb\nc\nd\ne\n", string(out))
}

func (s *ComplianceTestSuite) TestLLen() {
	cmd := exec.Command("redis-cli", "RPUSH", "list_key", "a", "b", "c", "d")
	out, err := cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("4\n", string(out))

	cmd = exec.Command("redis-cli", "LLEN", "list_key")
	out, err = cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("4\n", string(out))
}

func (s *ComplianceTestSuite) TestLLenNonExistentList() {
	cmd := exec.Command("redis-cli", "LLEN", "missing_key")
	out, err := cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("0\n", string(out))
}

func (s *ComplianceTestSuite) TestLPop() {
	cmd := exec.Command("redis-cli", "RPUSH", "list_key", "one", "two", "three", "four", "five")
	out, err := cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("5\n", string(out))

	cmd = exec.Command("redis-cli", "LPOP", "list_key")
	out, err = cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("one\n", string(out))

	// The popped element should be gone; remaining elements stay in order.
	cmd = exec.Command("redis-cli", "LRANGE", "list_key", "0", "-1")
	out, err = cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("two\nthree\nfour\nfive\n", string(out))
}

func (s *ComplianceTestSuite) TestLPopNonExistentList() {
	cmd := exec.Command("redis-cli", "LPOP", "missing_key")
	out, err := cmd.CombinedOutput()
	s.Require().NoError(err)
	// redis-cli renders a null bulk string ($-1\r\n) as a single newline.
	s.Equal("\n", string(out))
}

func (s *ComplianceTestSuite) TestLPopMultipleElements() {
	cmd := exec.Command("redis-cli", "RPUSH", "list_key", "one", "two", "three", "four", "five")
	out, err := cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("5\n", string(out))

	// LPOP with a count returns a RESP array of the removed elements.
	cmd = exec.Command("redis-cli", "LPOP", "list_key", "2")
	out, err = cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("one\ntwo\n", string(out))

	// Verify the remaining elements stay in their original order.
	cmd = exec.Command("redis-cli", "LRANGE", "list_key", "0", "-1")
	out, err = cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("three\nfour\nfive\n", string(out))
}

func (s *ComplianceTestSuite) TestLPopCountGreaterThanLength() {
	cmd := exec.Command("redis-cli", "RPUSH", "list_key", "a", "b", "c")
	out, err := cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("3\n", string(out))

	// Count exceeds length: remove and return every element.
	cmd = exec.Command("redis-cli", "LPOP", "list_key", "10")
	out, err = cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("a\nb\nc\n", string(out))

	// The list should now be empty.
	cmd = exec.Command("redis-cli", "LRANGE", "list_key", "0", "-1")
	out, err = cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("\n", string(out))
}

func (s *ComplianceTestSuite) TestBLPop() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	type result struct {
		out string
		err error
	}
	resCh := make(chan result, 1)

	go func() {
		cmd := exec.CommandContext(ctx, "redis-cli", "BLPOP", "list_key", "0")
		out, err := cmd.CombinedOutput()
		resCh <- result{string(out), err}
	}()

	// Give the BLPOP client time to connect and start blocking before we push.
	time.Sleep(100 * time.Millisecond)

	cmd := exec.CommandContext(ctx, "redis-cli", "RPUSH", "list_key", "foo")
	out, err := cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("1\n", string(out))

	select {
	case res := <-resCh:
		s.Require().NoError(res.err)
		// redis-cli renders the RESP array ["list_key", "foo"] as two lines.
		s.Equal("list_key\nfoo\n", res.out)
	case <-ctx.Done():
		s.T().Fatal("BLPOP did not return after RPUSH")
	}
}

func (s *ComplianceTestSuite) TestBLPopMultipleClientsServedInOrder() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	type result struct {
		out string
		err error
	}
	res1 := make(chan result, 1)
	res2 := make(chan result, 1)

	go func() {
		cmd := exec.CommandContext(ctx, "redis-cli", "BLPOP", "list_key", "0")
		out, err := cmd.CombinedOutput()
		res1 <- result{string(out), err}
	}()

	// Ensure client 1 is blocking before client 2 connects, so FIFO order is well-defined.
	time.Sleep(100 * time.Millisecond)

	go func() {
		cmd := exec.CommandContext(ctx, "redis-cli", "BLPOP", "list_key", "0")
		out, err := cmd.CombinedOutput()
		res2 <- result{string(out), err}
	}()

	time.Sleep(100 * time.Millisecond)

	// First push should wake the client that has been waiting the longest (client 1).
	cmd := exec.CommandContext(ctx, "redis-cli", "RPUSH", "list_key", "first")
	_, err := cmd.CombinedOutput()
	s.Require().NoError(err)

	select {
	case r := <-res1:
		s.Require().NoError(r.err)
		s.Equal("list_key\nfirst\n", r.out)
	case <-ctx.Done():
		s.T().Fatal("client 1 did not receive first push")
	}

	// Client 2 must still be blocked - a single push wakes only one waiter.
	select {
	case r := <-res2:
		s.T().Fatalf("client 2 returned before second push: %q", r.out)
	case <-time.After(100 * time.Millisecond):
	}

	cmd = exec.CommandContext(ctx, "redis-cli", "RPUSH", "list_key", "second")
	_, err = cmd.CombinedOutput()
	s.Require().NoError(err)

	select {
	case r := <-res2:
		s.Require().NoError(r.err)
		s.Equal("list_key\nsecond\n", r.out)
	case <-ctx.Done():
		s.T().Fatal("client 2 did not receive second push")
	}
}

func (s *ComplianceTestSuite) TestBLPopTimeoutExpires() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	start := time.Now()
	cmd := exec.CommandContext(ctx, "redis-cli", "BLPOP", "missing_key", "0.1")
	out, err := cmd.CombinedOutput()
	elapsed := time.Since(start)

	s.Require().NoError(err)
	// redis-cli renders a null array (*-1\r\n) as a single newline.
	s.Equal("\n", string(out))
	// Sanity check: the command must have actually blocked for ~the timeout duration.
	s.GreaterOrEqual(elapsed, 100*time.Millisecond)
}

func (s *ComplianceTestSuite) TestBLPopUnblocksBeforeTimeout() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	type result struct {
		out string
		err error
	}
	resCh := make(chan result, 1)

	go func() {
		// Use a generous timeout so the push has time to arrive first.
		cmd := exec.CommandContext(ctx, "redis-cli", "BLPOP", "list_key", "2")
		out, err := cmd.CombinedOutput()
		resCh <- result{string(out), err}
	}()

	time.Sleep(100 * time.Millisecond)

	cmd := exec.CommandContext(ctx, "redis-cli", "RPUSH", "list_key", "foo")
	_, err := cmd.CombinedOutput()
	s.Require().NoError(err)

	select {
	case res := <-resCh:
		s.Require().NoError(res.err)
		s.Equal("list_key\nfoo\n", res.out)
	case <-ctx.Done():
		s.T().Fatal("BLPOP did not return after RPUSH")
	}
}

func (s *ComplianceTestSuite) TestSetWithExpiry() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "redis-cli", "SET", "foo", "bar", "PX", "100")
	out, err := cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("OK\n", string(out))

	cmd = exec.CommandContext(ctx, "redis-cli", "GET", "foo")
	out, err = cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("bar\n", string(out))

	time.Sleep(200 * time.Millisecond)

	cmd = exec.CommandContext(ctx, "redis-cli", "GET", "foo")
	out, err = cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Equal("\n", string(out))
}
