package dkvs

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

// Starts a new master node; test reading, writing and getting the nodes list
func TestMaster(t *testing.T) {
	masterAddr := ":1212"
	encoding := "application/json"

	if m, err := NewMaster(masterAddr); m != nil {
		defer m.Close()
	} else if err != nil {
		t.Errorf("error creating master: %v", err)
		return
	}

	// wait for the server to start
	time.Sleep(100 * time.Millisecond)

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

	// Read
	url = "http://" + masterAddr + "/read"

	for k, v := range data {
		payload := map[string]string{
			"key": k,
		}
		jsonPayload, _ := json.Marshal(payload)
		buffer := bytes.NewBuffer(jsonPayload)

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

		if string(jsonVal) != v {
			t.Errorf("expected \"%s\", got \"%s\"", v, string(jsonVal))
		}
	}

	// multiple read
	url = "http://" + masterAddr + "/multi"

	p := `{"keys": [`
	for k := range data {
		p += `"` + k + `",`
	}
	// remove trailing , and close the payload
	p = p[:len(p)-1] + `]}`

	type responsePayload struct {
		Key   string `json:"k"`
		Value string `json:"v"`
		Error error  `json:"e"`
	}
	var rp []responsePayload

	buffer := bytes.NewBuffer([]byte(p))

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

	if len(rp) != len(data) {
		t.Errorf("expected a response container %d values, got %d instead", len(rp), len(data))
		return
	}

	hasError := false
	for _, val := range rp {
		if val.Error != nil {
			t.Errorf("missing value: %v", val.Error)
			hasError = true
		} else if expected := data[val.Key]; expected != val.Value {
			t.Errorf("expected value %v for key %v, got %v instead", expected, val.Key, val.Value)
			hasError = true
		}
	}

	if hasError {
		return
	}

	// list
	url = "http://" + masterAddr + "/list"

	payload := map[string]string{}
	jsonPayload, _ := json.Marshal(payload)
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

	decoder = json.NewDecoder(listResp.Body)
	if err = decoder.Decode(&list); err != nil {
		t.Errorf("error decoding the /list response: %v", err)
		return
	}

	if len(list) != 1 {
		t.Errorf("expected 1 node in the list, got %d", len(list))
		return
	}

	if !list[0].IsMaster() {
		t.Error("expected the node to be master")
		return
	}

	if list[0].Address != masterAddr {
		t.Errorf("expected address %s, got %s", masterAddr, list[0].Address)
		return
	}
}
