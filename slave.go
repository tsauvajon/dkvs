package dkvs

func (n *Node) checkMasterHealth() error {
	return errorNotImplemented
}

func (n *Node) electNewLeader() error {
	return errorNotImplemented
}

func (n *Node) promoteToMaster() error {
	return errorNotImplemented
}

// NewSlave creates a new node that joins an existing master
func NewSlave(addr, master string) (*Node, error) {
	n, err := newNode(addr)
	// TODO: register with the master
	return n, err
}
