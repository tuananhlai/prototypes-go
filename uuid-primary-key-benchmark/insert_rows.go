package uuidprimarykeybenchmark

import (
	"fmt"
	"math/rand/v2"
	"strings"
)

func InsertOneMilRows() error {
	rnd := rand.New(rand.NewChaCha8([32]byte{}))

	totalSize := 102
	batchSize := 10

	var valueStrParts []string
	var args []int

	for totalSize > 0 {
		sz := min(batchSize, totalSize)
		valueStrParts = valueStrParts[:0]
		args = args[:0]
		for range sz {
			valueStrParts = append(valueStrParts, "(?)")
			args = append(args, rnd.IntN(200))
		}

		valueStr := strings.Join(valueStrParts, ",")
		// _, err := db.Exec("INSERT INTO users (age) VALUES "+valueStr, args)
		// if err != nil {
		// 	return err
		// }
		fmt.Println(valueStr, args)

		totalSize -= sz
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
