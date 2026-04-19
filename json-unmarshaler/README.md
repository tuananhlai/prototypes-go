# Coding Challenge JSON Parser

A minimal JSON Parser written in Go.

## Examples

Read from STDIN.
```sh
$ echo '{
  "key1": true,
  "key2": false,
  "key3": null,
  "key4": "value",
  "key5": 101
}' | go run main.go
JSON is valid. ✅
```

Read from the given file.
```sh
$ go run main.go json/testdata/fail31.json
error tokenizing input: invalid number found at position 1: no digit found after exponent
exit status 1
```