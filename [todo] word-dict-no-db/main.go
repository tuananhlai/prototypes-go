package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

var wordToMeaning = map[string]string{
	"CABBLING": "The process of breaking up the flat masses into which wrought iron is first hammered, in order that the pieces may be reheated and wrought into bar iron.",
	"CABEZON":  "A California fish (Hemilepidotus spinosus), allied to the sculpin.",
}

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("invalid number of arguments: expect 1, got %d\n", len(os.Args)-1)
	}

	word := strings.ToUpper(os.Args[1])
	meaning, ok := wordToMeaning[word]
	if !ok {
		log.Fatalf("looking up %s: no meaning found", word)
	}

	fmt.Println(meaning)
}
