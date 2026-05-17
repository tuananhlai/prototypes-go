package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/tuananhlai/prototypes/my-redis/resp"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	defer l.Close()

	fmt.Println("Start server on port 6379")
	err = run(l)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(l net.Listener) error {
	executer := newExecuter(newStore())

	for {
		// TODO: add connection timeout.
		conn, err := l.Accept()
		if err != nil {
			return fmt.Errorf("accepting connection: %v", err)
		}

		go handleConn(conn, executer)
	}
}

func handleConn(conn net.Conn, executer *executer) {
	defer conn.Close()

	bufrw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	var res []byte

	for {
		// A Redis command will always be an non-empty array, with the first argument
		// being the command name.
		cmd, err := resp.ParseArray(bufrw.Reader)
		if err != nil {
			if err == io.EOF {
				return
			}

			res = resp.SerializeSimpleError(err.Error())
		} else {
			res = executer.execute(cmd)
		}

		_, err = bufrw.Write(res)
		if err != nil {
			panic(err)
		}

		err = bufrw.Flush()
		if err != nil {
			panic(err)
		}
	}
}

type entry struct {
	val       []byte
	expiredAt *time.Time
}

type store struct {
	mp map[string]entry
	mu sync.RWMutex
}

func newStore() *store {
	return &store{
		mp: make(map[string]entry),
	}
}

func (s *store) set(key string, val []byte, options *setOptions) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ent := entry{
		val:       val,
		expiredAt: options.expiredAt,
	}
	s.mp[key] = ent
}

func (s *store) get(key string) ([]byte, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	e, ok := s.mp[key]
	if !ok {
		return nil, false
	}
	if e.expiredAt != nil && time.Now().After(*e.expiredAt) {
		return nil, false
	}
	return e.val, ok
}

type setOptions struct {
	expiredAt *time.Time
}
