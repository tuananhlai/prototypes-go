package main

import (
	"fmt"

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

	name := string(cmd[0])
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
		if len(args) != 2 {
			return resp.SerializeSimpleError(fmt.Sprintf(
				"invalid number of arguments: expect 2, got %d", len(args)))
		}

		key, val := string(args[0]), args[1]
		ex.store.set(key, val)

		return resp.SerializeSimpleString("OK")
	default:
		return resp.SerializeSimpleError(fmt.Sprintf("unsupported command: %s", name))
	}
}
