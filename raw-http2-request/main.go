package main

import (
	"fmt"
	"log"

	"github.com/tuananhlai/prototypes/raw-http2-request/http2"
)

func main() {
	client := &http2.Client{}

	res, err := client.Get("https://example.com")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res)
}
