package dkvs

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

// Node is an autonomous kvs node that can be either a slave or a master
type Node struct {
	ID       string `json:"id"`
	MasterID string `json:"master"`
	Address  string `json:"addr"`

	nodes  map[string]*Node
	nMutex sync.RWMutex

	storage   Storage
	transport Transport
}

// ReadValue searches the value for the provided key in the storage
func (n *Node) ReadValue(key string) ([]byte, error) {
	return n.storage.Get(key)
}

// ReadMultipleValues searches for values associated with a range of keys
func (n *Node) ReadMultipleValues(keys ...string) ([]byte, error) {
	type payload struct {
		Key   string `json:"k"`
		Value string `json:"v"`
		Error error  `json:"e"`
	}
	p := make([]*payload, 0)

	for _, k := range keys {
		v, err := n.storage.Get(k)
		p = append(p, &payload{
			Key:   k,
			Value: string(v),
			Error: err,
		})
	}

	return json.Marshal(p)
}

// ListNodes returns a slice of all nodes
func (n *Node) ListNodes() ([]*Node, error) {
	// if the node list is empty (for example, in a slave that just got started)
	if n.nodes == nil {
		// we fetch the list from the master

	}

	nodes := make([]*Node, 0)
	n.nMutex.Lock()
	defer n.nMutex.Unlock()

	for _, node := range n.nodes {
		nodes = append(nodes, node)
	}
	return nodes, nil
}

// IsMaster checks if the current node is the master
func (n *Node) IsMaster() bool {
	return n.MasterID == n.ID
}

const (
	charBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

var _ = rand.NewSource(time.Now().UnixNano())

// generates a random ID
func newID(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = charBytes[rand.Intn(len(charBytes))]
	}
	return string(b)
}

func newNode(addr string) (*Node, error) {
	id := newID(16)

	n := &Node{
		ID:        id,
		nodes:     make(map[string]*Node),
		Address:   addr,
		storage:   NewStore(),
		transport: NewHTTPTransport(),
	}

	go func() {
		err := n.transport.Start(n)
		if err != nil {
			panic(fmt.Sprintf("failed to start transport with error: %v", err))
		}
	}()

	log.Printf("Created node with id %s\n", n.ID)

	return n, nil
}

// Close properly closes the node
func (n *Node) Close() error {
	// todo: send a message to master indicating that the node shut down
	return n.transport.Stop()
}
