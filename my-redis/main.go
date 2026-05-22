package main

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"slices"
	"time"

	"github.com/tuananhlai/prototypes/my-redis/resp"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to bind to port 6379")
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
	executor := newExecutor(newStore())

	for {
		// TODO: add connection timeout + limit number of concurrent clients.
		conn, err := l.Accept()
		if err != nil {
			return fmt.Errorf("accepting connection: %v", err)
		}

		go handleConn(conn, executor)
	}
}

func handleConn(conn net.Conn, executor *executor) {
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
			res = executor.execute(cmd)
		}

		_, err = bufrw.Write(res)
		if err != nil {
			slog.Error("write to buffer failed", "err", err)
			return
		}

		err = bufrw.Flush()
		if err != nil {
			slog.Error("flush buffer to client connection failed", "err", err)
			return
		}
	}
}

type entry struct {
	val       any
	expiredAt time.Time
}

type store struct {
	mp map[string]entry
}

func newStore() *store {
	return &store{
		mp: make(map[string]entry),
	}
}

func (s *store) set(key string, val []byte, expiredAt time.Time) error {
	rawExistingVal, keyAlreadyExists := s.mp[key]
	if keyAlreadyExists {
		if _, isValueDataTypeCorrect := rawExistingVal.val.([]byte); !isValueDataTypeCorrect {
			return fmt.Errorf("set key %s: wrong data type for existing value", key)
		}
	}

	s.mp[key] = entry{
		val:       val,
		expiredAt: expiredAt,
	}

	return nil
}

func (s *store) get(key string) ([]byte, error) {
	e, ok := s.mp[key]
	if !ok {
		return nil, nil
	}

	val, ok := e.val.([]byte)
	if !ok {
		return nil, fmt.Errorf("get key %s: invalid data type for value", key)
	}

	if !e.expiredAt.IsZero() && time.Now().After(e.expiredAt) {
		delete(s.mp, key)
		return nil, nil
	}

	return val, nil
}

func (s *store) rpush(key string, newElems [][]byte) (int, error) {
	var updatedVal [][]byte

	rawExistingVal, keyAlreadyExists := s.mp[key]
	if keyAlreadyExists {
		existingVal, isValueDataTypeCorrect := rawExistingVal.val.([][]byte)
		if !isValueDataTypeCorrect {
			return 0, fmt.Errorf("rpush %s: wrong data type for existing value", key)
		}
		updatedVal = existingVal
	}

	updatedVal = slices.Concat(updatedVal, newElems)
	s.mp[key] = entry{
		val: updatedVal,
	}

	return len(updatedVal), nil
}

func (s *store) lrange(key string, start, stop int) ([][]byte, error) {
	rawExistingVal, keyAlreadyExists := s.mp[key]
	if !keyAlreadyExists {
		return nil, nil
	}

	existingVal, isValueDataTypeCorrect := rawExistingVal.val.([][]byte)
	if !isValueDataTypeCorrect {
		return nil, fmt.Errorf("lrange %s: wrong data type for existing value", key)
	}

	if start >= len(existingVal) || start > stop {
		return nil, nil
	}

	stop = min(stop, len(existingVal)-1)
	return existingVal[start : stop+1], nil
}
