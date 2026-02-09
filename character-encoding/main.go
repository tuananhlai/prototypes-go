package main

import "fmt"

func main() {
	latin1 := []byte{0x63, 0x61, 0x66, 0xE9}
	fmt.Println(string(latin1), decodeLatin1(latin1)) // caf� café
	fmt.Printf("%v,%v", []byte("café"), []rune("café"))
}

// decodeLatin1 decodes the given bytes into UTF-8 assuming it is Latin1-encoded (0x00–0xFF).
func decodeLatin1(b []byte) string {
	runes := make([]rune, len(b))
	for i, c := range b {
		runes[i] = rune(c)
	}
	return string(runes)
}

// TODO: add shiftjis decoder
