package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	fmt.Println("Start server on port 6379")
	err = run(l)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			return fmt.Errorf("accepting connection: %v", err)
		}

		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	bufrw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	for {
		cmd, err := parse(bufrw.Reader)
		if err != nil {
			if err == io.EOF {
				return
			}
			panic(err)
		}

		res, err := execute(cmd)
		if err != nil {
			fmt.Printf("error executing command: %v\n", err)
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

func execute(cmd [][]byte) ([]byte, error) {
	if len(cmd) == 0 {
		return nil, errors.New("error empty command")
	}
	name := strings.ToUpper(string(cmd[0]))
	args := cmd[1:]

	executer, ok := commands[name]
	if !ok {
		return nil, fmt.Errorf("error unsupported command name: %v", name)
	}

	return executer(args)
}

type executer func(args [][]byte) (out []byte, err error)

var store = map[string][]byte{}

var commands = map[string]executer{
	"PING": func(args [][]byte) (out []byte, err error) {
		return []byte("+PONG\r\n"), nil
	},
	"ECHO": func(args [][]byte) (out []byte, err error) {
		if len(args) == 0 {
			return nil, errors.New("error missing argument")
		}
		return serializeBulkString(args[0])
	},
	"SET": func(args [][]byte) (out []byte, err error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("error invalid number of arguments: expect 2, got %d", len(args))
		}

		key, val := string(args[0]), args[1]
		store[key] = val

		return serializeSimpleString("OK")
	},
	"GET": func(args [][]byte) (out []byte, err error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("error invalid number of arguments: expect 1, got %d", len(args))
		}

		key := string(args[0])
		val, ok := store[key]
		if !ok {
			return serializeBulkString(nil)
		}

		return serializeBulkString(val)
	},
}
