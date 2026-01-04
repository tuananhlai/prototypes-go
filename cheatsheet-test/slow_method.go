package cheatsheettest

import "time"

func SlowMethod() string {
	time.Sleep(2 * time.Second)
	return "slow method"
}
