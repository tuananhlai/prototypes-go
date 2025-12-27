package main

import "time"

func main() {
	// read: can read from either master or replica.
	// write: only write to master.
	// because the data is replicated asynchronously, there's a chance that a replica
	// will have stale data.
}

type Main struct {
	data map[string]string
}

type Replica struct {
	data        map[string]string
	syncLatency time.Duration
}
