package dkvs

import "github.com/rs/xid"

// Node is an autonomous kvs node that can be either a slave or a master
type Node struct {
	id      string
	storage Storage
	// node list
	// transaction list
	// transaction log
}

func (n *Node) ReadValue(key string) ([]byte, error) {
	return n.storage.Get(key)
}

func NewNode() *Node {
	// todo :
	// instantiate as new master OR join an existing master
	guid := xid.New()
	n := &Node{
		id: guid.String(),
	}
	return n
}
