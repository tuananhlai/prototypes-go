package main

import (
	"bytes"
	"net"
	"os/exec"
	"sync"
	"testing"

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
