package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	data, _ := os.ReadFile("testdata/pass2.json")
	var result [][][][][][][][][][][][][][][][][][][]string
	err := json.Unmarshal(data, &result)
	fmt.Println(err, result)
}
