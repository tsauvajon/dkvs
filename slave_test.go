package dkvs

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

// Test replication (of writes and node list) to multiple slaves when writing
// to master
func TestSlave(t *testing.T) {
	masterAddr := ":2121"
	slaveAddr1 := ":2222"
	slaveAddr2 := ":2323"

	m, err := NewMaster(masterAddr)
	if m != nil {
		defer m.Close()
	}

	if err != nil {
		t.Errorf("creating a slave failed with error: %v", err)
		return
	}

	if m.ID == "" {
		t.Error("created nodes should have an id")
		return
	}

	// wait for the server to start
	time.Sleep(500 * time.Millisecond)

	s1, err := NewSlave(slaveAddr1, masterAddr)
	if s1 != nil {
		defer s1.Close()
	}
	if err != nil {
		t.Errorf("creating a slave failed with error: %v", err)
		return
	}

	s2, err := NewSlave(slaveAddr2, masterAddr)
	if s2 != nil {
		defer s2.Close()
	}
	if err != nil {
		t.Errorf("creating a slave failed with error: %v", err)
		return
	}

	time.Sleep(500 * time.Millisecond)

	// Write
	url := "http://" + masterAddr + "/write"

	payload := map[string]string{
		"key": "toto1",
		"val": "le 100",
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

	time.Sleep(500 * time.Millisecond)

	// get nodes list from slave 1
	url = "http://" + slaveAddr1 + "/list"
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
		buf := new(bytes.Buffer)
		buf.ReadFrom(listResp.Body)
		body := buf.String()
		t.Errorf("/list query failed with status %d: %v", listResp.StatusCode, body)
		return
	}

	var list []*Node

	decoder := json.NewDecoder(listResp.Body)
	if err = decoder.Decode(&list); err != nil {
		t.Errorf("error decoding the /list response: %v", err)
		return
	}

	if len(list) != 3 {
		t.Errorf("expected 3 nodes in the list, got %d", len(list))
		return
	}

	foundMaster := false

	for _, node := range list {
		if node.IsMaster() {
			if foundMaster {
				t.Error("expected only 1 master!")
				return
			}
			foundMaster = true
		}
	}

	if !foundMaster {
		t.Error("there is no master in this list!")
		return
	}

	// Read from slave 2
	url = "http://" + slaveAddr2 + "/read"

	payload = map[string]string{
		"key": "toto1",
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

	if string(jsonVal) != "le 100" {
		t.Errorf("expected \"le 100\", got \"%s\"", string(jsonVal[0:len(jsonVal)]))
	}
}

// Test initial replication
func TestInitialReplication(t *testing.T) {
	masterAddr := ":9873"
	slaveAddr := ":9872"

	m, err := NewMaster(masterAddr)
	if m != nil {
		defer m.Close()
	}

	if err != nil {
		t.Errorf("creating a slave failed with error: %v", err)
		return
	}

	if m.ID == "" {
		t.Error("created nodes should have an id")
		return
	}

	// wait for the server to start
	time.Sleep(500 * time.Millisecond)

	// Write
	url := "http://" + masterAddr + "/write"
	data := map[string]string{
		"toto":             "le sang",
		"qwerty":           "uiop",
		"toasdfg":          "hjkl",
		"zxcv":             "bnm",
		"qazwsxedffsdfs":   "salut",
		"pain au chocolat": "chocolatine",
	}

	for k, v := range data {
		payload := map[string]string{
			"key": k,
			"val": v,
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
	}

	time.Sleep(500 * time.Millisecond)

	slave, err := NewSlave(slaveAddr, masterAddr)
	if slave != nil {
		defer slave.Close()
	}
	if err != nil {
		t.Errorf("creating a slave failed with error: %v", err)
		return
	}

	// wait for the replication to complete
	time.Sleep(2000 * time.Millisecond)

	// Update a value
	url = "http://" + masterAddr + "/write"

	payload := map[string]string{
		"key": "toto",
		"val": "__le__100__",
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

	time.Sleep(500 * time.Millisecond)

	// Read all values
	url = "http://" + slaveAddr + "/multi"

	p := `{ "keys": [ "qwerty", "pain au chocolat", "toto" ]}`

	type responsePayload struct {
		Key   string `json:"k"`
		Value string `json:"v"`
		Error error  `json:"e"`
	}
	var rp []responsePayload

	buffer = bytes.NewBuffer([]byte(p))

	multiResp, err := http.Post(url, encoding, buffer)
	if err != nil {
		t.Errorf("error posting /multi: %v", err)
		return
	}
	defer multiResp.Body.Close()

	if multiResp.StatusCode != 200 {
		t.Errorf("/multi query failed with status code %d", multiResp.StatusCode)
		return
	}

	decoder := json.NewDecoder(multiResp.Body)
	if err = decoder.Decode(&rp); err != nil {
		t.Errorf("couldn't decode /multi response: %v", err)
		return
	}

	actualData := make(map[string]responsePayload)
	for _, payload := range rp {
		actualData[payload.Key] = payload
	}

	key := "pain au chocolat"
	if actual, expected := actualData[key].Value, data[key]; actual != expected {
		t.Errorf("expected value for %s to be %s, got %s", key, expected, actual)
		return
	}

	key = "qwerty"
	if actual, expected := actualData[key].Value, data[key]; actual != expected {
		t.Errorf("expected value for %s to be %s, got %s", key, expected, actual)
		return
	}

	key = "toto"
	if actual, expected := actualData[key].Value, "__le__100__"; actual != expected {
		t.Errorf("expected value for %s to be %s, got %s", key, expected, actual)
		return
	}
}

// Test that slaves cannot write
func TestWriteSlave(t *testing.T) {
	masterAddr := ":3121"
	slaveAddr := ":3222"

	m, err := NewMaster(masterAddr)
	if m != nil {
		defer m.Close()
	}

	if err != nil {
		t.Errorf("creating a slave failed with error: %v", err)
		return
	}

	if m.ID == "" {
		t.Error("created nodes should have an id")
		return
	}

	// wait for the server to start
	time.Sleep(500 * time.Millisecond)

	s, err := NewSlave(slaveAddr, masterAddr)
	if s != nil {
		defer s.Close()
	}
	if err != nil {
		t.Errorf("creating a slave failed with error: %v", err)
		return
	}

	time.Sleep(500 * time.Millisecond)

	// Write
	url := "http://" + slaveAddr + "/write"

	payload := map[string]string{
		"key": "qwerty",
		"val": "uiop",
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

	if resp.StatusCode == 200 || body != errorNotMaster.Error() {
		t.Errorf("should be denied, instead got: %v", body)
		return
	}
}
