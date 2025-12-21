package main

import "log"

type customError struct{}

func (c *customError) Error() string {
	return "custom error"
}

func doThing() *customError {
	return nil
}

// This program will panic instead of finishing successfully, even though
// `failed()` returns a `nil` pointer to `customError`. Why would that be?
func main() {
	var err error
	if err = doThing(); err != nil {
		log.Fatal(err)
	}

	log.Println("all good")
}
