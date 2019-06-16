package dkvs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func (n *Node) checkMasterHealth() error {
	return errorNotImplemented
}

func (n *Node) electNewLeader() (*Node, error) {
	return nil, errorNotImplemented
}

func (n *Node) promoteToMaster() error {
	return errorNotImplemented
}

// ReceiveListUpdate applies a nodes list update sent from the master
func (n *Node) ReceiveListUpdate(nodes map[string]*Node) error {
	n.nMutex.Lock()
	defer n.nMutex.Unlock()
	n.nodes = nodes

	return nil
}

// ReceiveWrite applies a write sent from the master
func (n *Node) ReceiveWrite(key, val string) error {
	if n.IsMaster() {
		return errorNotSlave
	}

	if err := n.storage.Set(key, val); err != nil {
		return err
	}

	log.Printf("node %s replicated key %s", n.ID, key)

	return nil
}

// NewSlave creates a new node that joins an existing master
func NewSlave(addr, master string) (*Node, error) {
	n, err := newNode(addr)

	url := "http://" + master + "/join"
	encoding := "application/json"
	payload, _ := json.Marshal(n)
	buffer := bytes.NewBuffer(payload)

	resp, err := http.Post(url, encoding, buffer)
	if err != nil {
		return nil, fmt.Errorf("joining master: %v", err)
	}

	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	body := buf.String()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("joining master bad response: %v", body)
	}

	// TODO: register with the master
	// TODO: set MasterID
	return n, err
}
