package main

import (
	"strconv"
	"strings"
)

func main() {
}

func serializeIntArrayStrconv(arr []int) []byte {
	arrStr := make([]string, 0, len(arr))

	for _, val := range arr {
		arrStr = append(arrStr, strconv.Itoa(val))
	}

	return []byte(strings.Join(arrStr, ","))
}

func deserializeIntArrayStrconv(data []byte) ([]int, error) {
	arrStr := strings.Split(string(data), ",")
	var err error
	arr := make([]int, len(arrStr))

	for i, val := range arrStr {
		arr[i], err = strconv.Atoi(val)
		if err != nil {
			return nil, err
		}
	}

	return arr, nil
}
