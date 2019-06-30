package dkvs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

	n.MasterID = nodes[n.ID].MasterID

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

// ReplicateFromMaster will read a stream of data from the master and save it
// locally to this slave
// next steps:
// 1/ until the initial replication is complete, slave shouldn't accept reads: it
// will either return an error or redirect the read query to another node.
// 2/ slaves will still accept writes but store them in an ordered queue. It
// will apply all the writes in sequential order (first in, first out) once the
// replication is done
func (n *Node) ReplicateFromMaster(r io.Reader) error {
	err := n.storage.ReplicateFrom(r)
	if err != nil {
		log.Println("shutting the node down because replication failed: ", err)
		defer n.Close()
	}

	log.Printf("node %s replicated the database", n.ID)

	return err
}

// NewSlave creates a new node that joins an existing master
func NewSlave(addr, master string) (*Node, error) {
	n, err := newNode(addr)

	if err != nil {
		defer n.Close()
		return nil, fmt.Errorf("creating node: %v", err)
	}

	url := "http://" + master + "/join"
	payload, _ := json.Marshal(n)
	buffer := bytes.NewBuffer(payload)

	resp, err := http.Post(url, encoding, buffer)
	if err != nil {
		defer n.Close()
		return nil, fmt.Errorf("joining master: %v", err)
	}

	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	body := buf.String()

	if resp.StatusCode != 200 {
		defer n.Close()
		return nil, fmt.Errorf("joining master bad response: %v", body)
	}

	return n, err
}
