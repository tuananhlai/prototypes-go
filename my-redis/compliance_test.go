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
