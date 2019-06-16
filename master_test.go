package dkvs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

// Starts a new master node; test reading, writing and getting the nodes list
func TestMaster(t *testing.T) {
	masterAddr := ":1212"

	if m, err := NewMaster(masterAddr); m != nil {
		defer m.Close()
	} else if err != nil {
		t.Errorf("error creating master: %v", err)
		return
	}

	// wait for the server to start
	time.Sleep(1 * time.Second)

	// Write
	url := "http://" + masterAddr + "/write"
	encoding := "application/json"

	payload := map[string]string{
		"key": "toto",
		"val": "le sang",
	}
	jsonPayload, _ := json.Marshal(payload)
	buffer := bytes.NewBuffer(jsonPayload)

	resp, err := http.Post(url, encoding, buffer)
	if err != nil {
		t.Errorf("error posting /write: %v", err)
		return
	}
	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	body := buf.String()

	if resp.StatusCode != 200 {
		t.Errorf("/write query failed: %v", body)
		return
	}

	// Read
	url = "http://" + masterAddr + "/read"
	encoding = "application/json"

	payload = map[string]string{
		"key": "toto",
	}
	jsonPayload, _ = json.Marshal(payload)
	buffer = bytes.NewBuffer(jsonPayload)

	readResp, err := http.Post(url, encoding, buffer)
	if err != nil {
		t.Errorf("error posting /read: %v", err)
		return
	}
	defer readResp.Body.Close()
	jsonVal, err := ioutil.ReadAll(readResp.Body)

	if err != nil {
		t.Errorf("couldn't decode /read response: %v", err)
		return
	}

	if readResp.StatusCode != 200 {
		t.Errorf("/read query failed: %v", string(jsonVal))
		return
	}

	if string(jsonVal) != "le sang" {
		t.Errorf("expected \"le sang\", got \"%s\"", string(jsonVal[0:len(jsonVal)]))
	}

	// list
	url = "http://" + masterAddr + "/list"
	encoding = "application/json"

	payload = map[string]string{}
	jsonPayload, _ = json.Marshal(payload)
	buffer = bytes.NewBuffer(jsonPayload)

	listResp, err := http.Post(url, encoding, buffer)
	if err != nil {
		t.Errorf("error posting /list: %v", err)
		return
	}
	defer listResp.Body.Close()

	if listResp.StatusCode != 200 {
		buf2 := new(bytes.Buffer)
		buf2.ReadFrom(listResp.Body)
		body = buf2.String()
		t.Errorf("/list query failed with status %d: %v", listResp.StatusCode, body)
		return
	}

	var list []*Node

	decoder := json.NewDecoder(listResp.Body)
	if err = decoder.Decode(&list); err != nil {
		fmt.Println("listresp.body: ", listResp.Body)
		fmt.Println("body: ", body)
		t.Errorf("error decoding the /list response: %v", err)
		return
	}

	if len(list) != 1 {
		t.Errorf("expected 1 node in the list, got %d", len(list))
		return
	}

	if !list[0].Master {
		t.Error("expected the node to be master")
		return
	}

	if list[0].Address != masterAddr {
		t.Errorf("expected address %s, got %s", masterAddr, list[0].Address)
		return
	}
}
