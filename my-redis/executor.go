package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/tuananhlai/prototypes/my-redis/resp"
)

type command struct {
	name  string
	args  [][]byte
	reply chan []byte
}

// executor parses and runs commands in a single thread.
type executor struct {
	store *store
	queue chan command
}

func newExecutor(store *store) *executor {
	ex := &executor{
		store: store,
		queue: make(chan command, 100),
	}
	go ex.loop()

	return ex
}

func (ex *executor) loop() {
	for cmd := range ex.queue {
		switch cmd.name {
		case "PING":
			cmd.reply <- resp.SerializeSimpleString("PONG")
		case "ECHO":
			if len(cmd.args) == 0 {
				cmd.reply <- resp.SerializeSimpleError("missing argument")
				return
			}
			cmd.reply <- resp.SerializeBulkString(cmd.args[0])
		case "GET":
			if len(cmd.args) != 1 {
				cmd.reply <- resp.SerializeSimpleError(fmt.Sprintf(
					"invalid number of arguments: expect 1, got %d", len(cmd.args)))
				return
			}

			key := string(cmd.args[0])
			val, err := ex.store.get(key)
			if err != nil {
				cmd.reply <- resp.SerializeSimpleError(err.Error())
				return
			}
			cmd.reply <- resp.SerializeBulkString(val)
		case "SET":
			setArgs, err := parseSetCmdArgs(cmd.args)
			if err != nil {
				cmd.reply <- resp.SerializeSimpleError(err.Error())
				return
			}

			err = ex.store.set(setArgs.key, setArgs.val, setArgs.expiredAt)
			if err != nil {
				cmd.reply <- resp.SerializeSimpleError(err.Error())
				return
			}
			cmd.reply <- resp.SerializeSimpleString("OK")
		case "RPUSH":
			if len(cmd.args) < 2 {
				cmd.reply <- resp.SerializeSimpleError(fmt.Sprintf(
					"invalid number of arguments: expect at least 2, got %d", len(cmd.args)))
				return
			}

			key := string(cmd.args[0])
			values := cmd.args[1:]

			cnt, err := ex.store.rpush(key, values)
			if err != nil {
				cmd.reply <- resp.SerializeSimpleError(err.Error())
				return
			}

			cmd.reply <- resp.SerializeInteger(cnt)
		case "LRANGE":
			if len(cmd.args) != 3 {
				cmd.reply <- resp.SerializeSimpleError(fmt.Sprintf(
					"invalid number of arguments: expect 3, got %d", len(cmd.args)))
				return
			}

			key := string(cmd.args[0])
			start, err := strconv.Atoi(string(cmd.args[1]))
			if err != nil {
				cmd.reply <- resp.SerializeSimpleError(err.Error())
				return
			}

			stop, err := strconv.Atoi(string(cmd.args[2]))
			if err != nil {
				cmd.reply <- resp.SerializeSimpleError(err.Error())
				return
			}

			retval, err := ex.store.lrange(key, start, stop)
			if err != nil {
				cmd.reply <- resp.SerializeSimpleError(err.Error())
				return
			}

			cmd.reply <- resp.SerializeArray(retval)
		default:
			cmd.reply <- resp.SerializeSimpleError(
				fmt.Sprintf("unsupported command: %s", cmd.name))
		}

		close(cmd.reply)
	}
}

// execute parses and runs the given command array and returns its
// output.
func (ex *executor) execute(rawCmd [][]byte) []byte {
	if len(rawCmd) == 0 {
		return resp.SerializeSimpleError("empty command")
	}

	name := strings.ToUpper(string(rawCmd[0]))
	args := rawCmd[1:]

	cmd := command{
		name:  name,
		args:  args,
		reply: make(chan []byte, 1),
	}
	ex.queue <- cmd

	return <-cmd.reply
}

type setCmdArgs struct {
	key       string
	val       []byte
	expiredAt time.Time
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

	key, eof := read()
	if eof {
		err = errors.New("reading `key`: unexpected EOF")
		return
	}
	retval.key = string(key)

	val, eof := read()
	if eof {
		err = errors.New("reading `val`: unexpected EOF")
		return
	}
	retval.val = val

	// TODO: parse according to Redis's syntax diagram.
	// https://redis.io/docs/latest/commands/set/
	for {
		// parse command options
		optNameBytes, eof := read()
		if eof {
			return
		}
		optName := strings.ToUpper(string(optNameBytes))

		switch optName {
		case "PX":
			expiryMillisBytes, eof := read()
			if eof {
				err = fmt.Errorf("reading value for option '%s': unexpected EOF", optNameBytes)
				return
			}

			var expiryMillis int
			expiryMillis, err = strconv.Atoi(string(expiryMillisBytes))
			if err != nil {
				err = fmt.Errorf("convert duration to int: %w", err)
				return
			}

			retval.expiredAt = time.Now().Add(time.Duration(expiryMillis) * time.Millisecond)
		default:
			err = fmt.Errorf("invalid option %s", optName)
			return
		}
	}
}
