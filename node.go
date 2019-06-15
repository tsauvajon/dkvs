package dkvs

import (
	"fmt"

	"github.com/rs/xid"
)

// Node is an autonomous kvs node that can be either a slave or a master
type Node struct {
	ID        string `json:"id"`
	storage   Storage
	transport Transport
	nodes     map[string]*Node
	Master    bool   `json:"master"`
	Address   string `json:"addr"`
	// transaction list
	// transaction log
}

// ReadValue will search the value for the provided key in the storage
func (n *Node) ReadValue(key string) ([]byte, error) {
	return n.storage.Get(key)
}

// NewMaster creates a new node as a master
func NewMaster(addr string) (*Node, error) {
	n, err := newNode(addr)
	n.Master = true
	n.nodes[n.ID] = n
	return n, err
}

// NewSlave creates a new node that joins an existing master
func NewSlave(addr, master string) (*Node, error) {
	n, err := newNode(addr)
	// TODO: register with the master
	return n, err
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
		err := n.transport.Start(n.Address)
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
