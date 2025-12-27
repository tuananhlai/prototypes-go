package main

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

// demonstrate how quorum read and write work. does not handle node errors
// or edge cases yet.
func main() {
	nodes := []*Node{
		NewNode("1", 200*time.Millisecond),
		NewNode("2", 200*time.Millisecond),
		NewNode("3", 3000*time.Millisecond),
	}
	cluster, err := NewCluster(nodes, 2)
	if err != nil {
		log.Fatal(err)
	}

	cluster.Write("foo", "fooValue")
	cluster.Write("bar", "barValue")
	cluster.Write("baz", "bazValue")

	log.Printf(
		`cluster.Read("foo") -> '%s', cluster.Read("bar") -> '%s', cluster.Read("baz") -> '%s'\n`,
		cluster.Read("foo"),
		cluster.Read("bar"),
		cluster.Read("baz"),
	)

	log.Printf(
		`cluster.ConsistentRead("foo") -> '%s', cluster.ConsistentRead("bar") -> '%s', cluster.ConsistentRead("baz") -> '%s'\n`,
		cluster.ConsistentRead("foo"),
		cluster.ConsistentRead("bar"),
		cluster.ConsistentRead("baz"),
	)

	log.Println(cluster.ConsistentRead("foo"), cluster.ConsistentRead("bar"), cluster.ConsistentRead("baz"))
}

type Cluster struct {
	nodes      []*Node
	readQuorum int
}

func NewCluster(nodes []*Node, readQuorum int) (*Cluster, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("nodes must not be empty")
	}
	if readQuorum <= 0 || readQuorum > len(nodes) {
		return nil, fmt.Errorf("readQuorum must be between 1 and %d", len(nodes))
	}

	return &Cluster{
		nodes:      nodes,
		readQuorum: readQuorum,
	}, nil
}

func (c *Cluster) Write(key, val string) {
	start := time.Now()
	ackCh := make(chan struct{}, len(c.nodes))
	for _, node := range c.nodes {
		go func(nde *Node) {
			nde.Write(key, Entry{
				Value: val,
				// generate the version on the coordinator to ensure monotonicity.
				Version: time.Now(),
			})
			ackCh <- struct{}{}
		}(node)
	}

	numAck := 0
	for _ = range ackCh {
		numAck++
		if numAck >= c.writeQuorum() {
			log.Printf("Write %s:%s. Took %s", key, val, time.Since(start))
			return
		}
	}
}

func (c *Cluster) Read(key string) string {
	start := time.Now()
	node := c.getRandomNode()
	val := node.Read(key).Value
	log.Printf("Read node %s. Returned '%s'. Took %s", node.id, val, time.Since(start))
	return val
}

func (c *Cluster) ConsistentRead(key string) string {
	start := time.Now()

	// Creates a buffered channel to allow all go routine to finish
	resultCh := make(chan Entry, len(c.nodes))
	for _, node := range c.nodes {
		go func(node *Node) {
			resultCh <- node.Read(key)
		}(node)
	}

	results := make([]Entry, 0)
	for v := range resultCh {
		results = append(results, v)
		if len(results) >= c.readQuorum {
			break
		}
	}

	latestResult := Entry{}
	for _, result := range results {
		if result.Version.After(latestResult.Version) {
			latestResult = result
		}
	}

	log.Printf("ConsistentRead returns '%s'. Took %s", latestResult.Value, time.Since(start))
	return latestResult.Value
}

func (c *Cluster) writeQuorum() int {
	return len(c.nodes) + 1 - c.readQuorum
}

func (c *Cluster) getRandomNode() *Node {
	return c.nodes[rand.Intn(len(c.nodes))]
}

type Entry struct {
	Value   string
	Version time.Time
}

type Node struct {
	id           string
	data         map[string]Entry
	mu           sync.RWMutex
	writeLatency time.Duration
}

func NewNode(id string, latency time.Duration) *Node {
	return &Node{
		id:           id,
		data:         map[string]Entry{},
		mu:           sync.RWMutex{},
		writeLatency: latency,
	}
}

func (node *Node) Write(key string, entry Entry) {
	time.Sleep(node.writeLatency)
	node.mu.Lock()
	node.data[key] = entry
	node.mu.Unlock()
	log.Printf("[Node %s] %s:%s written\n", node.id, key, entry.Value)
}

func (node *Node) Read(key string) Entry {
	time.Sleep(100 * time.Millisecond)
	node.mu.RLock()
	defer node.mu.RUnlock()
	entry, _ := node.data[key]
	return entry
}
