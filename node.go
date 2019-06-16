package dkvs

import (
	"fmt"
	"sync"

	"github.com/rs/xid"
)

// Node is an autonomous kvs node that can be either a slave or a master
type Node struct {
	ID      string `json:"id"`
	Master  bool   `json:"master"`
	Address string `json:"addr"`

	nodes  map[string]*Node
	nMutex sync.RWMutex

	storage   Storage
	transport Transport
}

// ReadValue searches the value for the provided key in the storage
func (n *Node) ReadValue(key string) ([]byte, error) {
	return n.storage.Get(key)
}

// ListNodes returns a slice of all nodes
func (n *Node) ListNodes() ([]*Node, error) {
	nodes := make([]*Node, 0)
	n.nMutex.Lock()
	defer n.nMutex.Unlock()

	for _, node := range n.nodes {
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func newNode(addr string) (*Node, error) {
	n := &Node{
		ID:        xid.New().String(),
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

	return n, nil
}

// Close will properly close the node
func (n *Node) Close() error {
	return n.transport.Stop()
}
