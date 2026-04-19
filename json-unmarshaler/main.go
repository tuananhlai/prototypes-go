package main

import (
	"fmt"
	"io"
	"os"

	ccjson "github.com/tuananhlai/prototypes-go/json-parser/json"
)

func main() {
	var jsonInput []byte
	var err error

	if len(os.Args) == 1 {
		jsonInput, err = io.ReadAll(os.Stdin)
	} else {
		jsonInput, err = os.ReadFile(os.Args[1])
	}

	if err != nil {
		fmt.Printf("error reading json input: %v\n", err)
		os.Exit(1)
	}

	_, err = ccjson.Parse(string(jsonInput))
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	fmt.Println("JSON is valid. ✅")
}
