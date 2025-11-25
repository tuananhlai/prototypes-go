package main

import (
	"fmt"
	"unicode/utf8"
)

func main() {
	s := "世界 means world 🌎"

	var buf [utf8.UTFMax]byte

	for i, r := range s {
		rl := utf8.RuneLen(r)
		si := i + rl
		copy(buf[:], s[i:si])

		fmt.Printf("%2d: %q; codepoint: %#6x; encoded bytes: %#v\n", i, r, r, buf[:rl])
	}
}
