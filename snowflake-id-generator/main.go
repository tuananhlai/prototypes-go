package main

import (
	"fmt"
	"log"

	"github.com/tuananhlai/prototypes/snowflake-id-generator/snowflake"
)

func main() {
	node, err := snowflake.NewNode(125)
	if err != nil {
		log.Fatal(err)
	}

	for range 10 {
		id := node.Generate()
		fmt.Printf("%d %x\n", id, id)
	}
}
