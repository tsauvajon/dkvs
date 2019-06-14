package dkvs

import "github.com/rs/xid"

// Node is an autonomous kvs node that can be either a slave or a master
type Node struct {
	id        string
	storage   Storage
	transport Transport
	allNodes  []*Node
	// transaction list
	// transaction log
}

// ReadValue will search the value for the provided key in the storage
func (n *Node) ReadValue(key string) ([]byte, error) {
	return n.storage.Get(key)
}

// NewMaster creates a new node as a master
func NewMaster() (*Node, error) {
	return newNode(), nil
}

// NewSlave creates a new node that joins an existing master
func NewSlave(master string) (*Node, error) {
	return nil, errorNotImplemented
}

func newNode() *Node {
	guid := xid.New()
	n := &Node{
		id: guid.String(),
	}
	return n
}
