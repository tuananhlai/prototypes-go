package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/tuananhlai/prototypes/my-redis/resp"
)

type executer struct {
	store *store
}

func newExecuter(store *store) *executer {
	return &executer{
		store: store,
	}
}

// execute parses and runs the given command array and returns its
// output.
func (ex *executer) execute(cmd [][]byte) []byte {
	if len(cmd) == 0 {
		return resp.SerializeSimpleError("empty command")
	}

	name := strings.ToUpper(string(cmd[0]))
	args := cmd[1:]
	switch name {
	case "PING":
		return resp.SerializeSimpleString("PONG")
	case "ECHO":
		if len(args) == 0 {
			return resp.SerializeSimpleError("missing argument")
		}
		res, err := resp.SerializeBulkString(args[0])
		if err != nil {
			return resp.SerializeSimpleError(err.Error())
		}

		return res
	case "GET":
		if len(args) != 1 {
			return resp.SerializeSimpleError(fmt.Sprintf(
				"invalid number of arguments: expect 1, got %d", len(args)))
		}

		key := string(args[0])
		val, ok := ex.store.get(key)
		if !ok {
			return resp.NullBulkString
		}

		res, err := resp.SerializeBulkString(val)
		if err != nil {
			return resp.SerializeSimpleError(err.Error())
		}

		return res
	case "SET":
		setArgs, err := parseSetCmdArgs(args)
		if err != nil {
			return resp.SerializeSimpleError(err.Error())
		}

		ex.store.set(setArgs.key, setArgs.val, &setOptions{
			expiredAt: setArgs.expiredAt,
		})

		return resp.SerializeSimpleString("OK")
	default:
		return resp.SerializeSimpleError(fmt.Sprintf("unsupported command: %s", name))
	}
}

type setCmdArgs struct {
	key       string
	val       []byte
	expiredAt *time.Time
}

func parseSetCmdArgs(args [][]byte) (retval setCmdArgs, err error) {
	cur := 0
	read := func() (retval []byte, isEOF bool) {
		if cur >= len(args) {
			isEOF = true
			return
		}

		retval = args[cur]
		cur++

		return
	}

	key, isEOF := read()
	if isEOF {
		err = errors.New("reading `key`: unexpected EOF")
		return
	}
	retval.key = string(key)

	val, isEOF := read()
	if isEOF {
		err = errors.New("reading `val`: unexpected EOF")
		return
	}
	retval.val = val

	// parse command options
	optNameBytes, isEOF := read()
	if isEOF {
		return
	}
	optName := strings.ToUpper(string(optNameBytes))

	// TODO: parse according to Redis's syntax diagram.
	// https://redis.io/docs/latest/commands/set/
	switch optName {
	case "PX":
		expiryMillisBytes, isEOF := read()
		if isEOF {
			err = fmt.Errorf("reading value for option '%s': unexpected EOF", optNameBytes)
			return
		}

		var expiryMillis int
		expiryMillis, err = strconv.Atoi(string(expiryMillisBytes))
		if err != nil {
			err = fmt.Errorf("convert duration to int: %w", err)
			return
		}

		expiredAt := time.Now().Add(time.Duration(expiryMillis) * time.Millisecond)
		retval.expiredAt = &expiredAt
	default:
		err = fmt.Errorf("invalid option %s", optName)
	}

	return
}
