package snowflake

import (
	"fmt"
	"sync"
	"time"
)

const (
	timeShift        = 22
	machineShift     = 12
	numMachineIDBits = 10
	numCounterBits   = 12
	maxMachineID     = (1 << numMachineIDBits) - 1
	counterMask      = (1 << numCounterBits) - 1
	maxCounter       = counterMask
)

type Node struct {
	machineID  uint64
	mu         sync.Mutex
	counter    uint64
	lastMillis int64
}

func NewNode(machineID int) (*Node, error) {
	if machineID < 0 || machineID > maxMachineID {
		return nil, fmt.Errorf("machine ID must be between 0 and %d", maxMachineID)
	}
	return &Node{
		machineID: uint64(machineID),
	}, nil
}

func (node *Node) Generate() uint64 {
	node.mu.Lock()
	defer node.mu.Unlock()

	timeMillis := time.Now().UnixMilli()

	// We need to use a loop for waiting and retrying logic once our counter (sequence) maxed out.
	for {
		if timeMillis == node.lastMillis && node.counter == maxCounter {
			time.Sleep(time.Microsecond)
			timeMillis = time.Now().UnixMilli()
			continue
		}

		if timeMillis == node.lastMillis {
			node.counter++
		} else {
			node.lastMillis = timeMillis
			node.counter = 0
		}
		break
	}

	timeMillisBits := uint64(timeMillis << timeShift)
	machineIDBits := node.machineID << machineShift
	counterBits := node.counter & counterMask

	return timeMillisBits | machineIDBits | counterBits
}
