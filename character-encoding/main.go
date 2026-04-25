package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

func main() {
	latin1 := []byte{0x63, 0x61, 0x66, 0xE9}
	fmt.Println(string(latin1), decodeLatin1(latin1)) // caf� café
	fmt.Printf("%v,%v\n", []byte("café"), []rune("café"))

	shiftjisStr := []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x2C, 0x20, 0x90, 0xA2, 0x8A, 0x45, 0x81, 0x49} // Hello, 世界！
	decodedShiftJISStr, err := decodeShiftJIS(shiftjisStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error decoding shift jis string: %v", err)
		return
	}
	fmt.Println(string(shiftjisStr), decodedShiftJISStr)
}

// decodeLatin1 decodes the given bytes into UTF-8 assuming it is Latin1-encoded (0x00–0xFF).
func decodeLatin1(b []byte) string {
	runes := make([]rune, len(b))
	for i, c := range b {
		runes[i] = rune(c)
	}
	return string(runes)
}

func decodeShiftJIS(b []byte) (string, error) {
	r := bytes.NewReader(b)

	var retval []rune
	for {
		curByte, err := r.ReadByte()
		if err != nil {
			if err == io.EOF {
				return string(retval), nil
			}

			return "", err
		}

		// ASCII characters
		if curByte <= 0x7F {
			retval = append(retval, rune(curByte))
			continue
		}

		// Half-width Katakana
		if curByte >= 0xA1 && curByte <= 0xDF {
			retval = append(retval, shiftJISByteToUTF8[uint16(curByte)])
			continue
		}

		// Full-width Kanji-Kana. If the current byte falls inside a certain range, we need to read the next
		// byte as well and decode the character using these two bytes
		if (curByte >= 0x81 && curByte <= 0x9F) || (curByte >= 0xE0 && curByte <= 0xEF) {
			trailByte, err := r.ReadByte()
			if err != nil {
				return "", err
			}
			retval = append(retval, shiftJISByteToUTF8[uint16(curByte)*256+uint16(trailByte)])
			continue
		}

		return "", fmt.Errorf("error unsupported byte found: 0x%x", curByte)
	}
}

var shiftJISByteToUTF8 = map[uint16]rune{
	0x90A2: '世',
	0x8A45: '界',
	0x8149: '！',
}
