package dkvs

import "fmt"

// Check if the nodes are healthy, update the state of all nodes.
func (n *Node) checkNodesHealth() error {
	return ERROR_NOT_IMPLEMENTED
}

// Replicates a write to all the nodes
func (n *Node) pushSetToNodes(key, val string) error {
	return ERROR_NOT_IMPLEMENTED
}

// Checks if this node is the master
func (n *Node) isMaster() (bool, error) {
	return false, ERROR_NOT_IMPLEMENTED
}

// WriteValue will write a value to the internal
// storage and push it to all the slaves.
// This can only be run on the master.
func (n *Node) WriteValue(key, val string) error {
	if m, err := n.isMaster(); err != nil {
		return fmt.Errorf("could not check if node is master: %v", err)
	} else if !m {
		return ERROR_NOT_MASTER
	}

	if err := n.storage.Set(key, val); err != nil {
		return err
	}

	return n.pushSetToNodes(key, val)
}
