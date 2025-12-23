package counter

import "sync"

func CountToFour() int {
	count := 0

	var wg sync.WaitGroup
	for range 2 {
		wg.Go(func() {
			for range 2 {
				count++
			}
		})
	}
	wg.Wait()

	return count
}
