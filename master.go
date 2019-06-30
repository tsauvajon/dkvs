package dkvs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Check if the nodes are healthy, update the state of all nodes.
func (n *Node) checkNodesHealth() error {
	return errorNotImplemented
}

// Replicates a write to all the nodes
func (n *Node) pushWriteToSlaves(key, val string) error {
	for id, slave := range n.nodes {
		// do not push to self
		if id == n.ID {
			continue
		}

		// run goroutines to asynchronously push to all slaves, and retry on fails
		go func(slave *Node) {
			for i := 0; i < defaultConfig.retriesCount; i++ {
				if err := n.pushWriteToOneSlave(slave, key, val); err != nil {
					log.Printf("try %d pushing to %s: %v", i+1, slave.ID, err)
					time.Sleep(defaultConfig.retriesDelayMs * time.Millisecond)
				}
				break
			}
		}(slave)
	}
	return nil
}

func (n *Node) pushWriteToOneSlave(slave *Node, key, val string) error {
	url := "http://" + slave.Address + "/receive"

	payload := map[string]string{
		"key": key,
		"val": val,
	}
	jsonPayload, _ := json.Marshal(payload)
	buffer := bytes.NewBuffer(jsonPayload)

	resp, err := http.Post(url, encoding, buffer)
	if err != nil {
		return fmt.Errorf("pushing write: %v", err)
	}
	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	body := buf.String()

	if resp.StatusCode != 200 {
		return fmt.Errorf("pushing write bad response: %v", body)
	}

	log.Printf("pushed write to %s", slave.ID)

	return nil
}

// Replicates a list update to all the nodes
func (n *Node) pushListUpdateToSlaves() error {
	for id, slave := range n.nodes {
		// do not push to self
		if id == n.ID {
			continue
		}

		// run goroutines to asynchronously push to all slaves, and retry on fails
		go func(slave *Node) {
			for i := 0; i < defaultConfig.retriesCount; i++ {
				if err := n.pushListUpdateToOneSlave(slave); err == nil {
					break
				}
				time.Sleep(defaultConfig.retriesDelayMs * time.Millisecond)
			}
		}(slave)
	}
	return nil
}

func (n *Node) pushListUpdateToOneSlave(slave *Node) error {
	url := "http://" + slave.Address + "/update"
	payload, _ := json.Marshal(n.nodes)
	buffer := bytes.NewBuffer(payload)

	resp, err := http.Post(url, encoding, buffer)
	if err != nil {
		return fmt.Errorf("pushing list update: %v", err)
	}

	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	body := buf.String()

	if resp.StatusCode != 200 {
		return fmt.Errorf("pushing list update bad response: %v", body)
	}

	return nil
}

// replicateToSlave replicates data by streaming it from the master to the slave.
func (n *Node) replicateToSlave(slave *Node) error {
	buffer, err := n.storage.ReplicateTo()

	if err != nil {
		return err
	}

	url := "http://" + slave.Address + "/replicate"

	resp, err := http.Post(url, encoding, buffer)
	if err != nil {
		return fmt.Errorf("replicate: %v", err)
	}
	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	body := buf.String()

	if resp.StatusCode != 200 {
		return fmt.Errorf("replicate bad response: %v", body)
	}

	return nil
}

// NewMaster creates a new node as a master
func NewMaster(addr string) (*Node, error) {
	n, err := newNode(addr)
	n.MasterID = n.ID
	n.nodes[n.ID] = n
	return n, err
}

// Join allows a slave to join this node
func (n *Node) Join(slave *Node) error {
	if !n.IsMaster() {
		return errorNotMaster
	}

	n.nMutex.Lock()
	defer n.nMutex.Unlock()

	slave.MasterID = n.MasterID
	n.nodes[slave.ID] = slave

	if err := n.replicateToSlave(slave); err != nil {
		return err
	}

	log.Printf("node %s joined", slave.ID)

	return n.pushListUpdateToSlaves()
}

// WriteValue will write a value to the internal
// storage and push it to all the slaves.
// This can only be run on the master.
func (n *Node) WriteValue(key, val string) error {
	if !n.IsMaster() {
		return errorNotMaster
	}

	if err := n.storage.Set(key, val); err != nil {
		return err
	}

	return n.pushWriteToSlaves(key, val)
}
