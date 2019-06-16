package dkvs

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestMaster(t *testing.T) {
	masterAddr := ":1212"

	// os.Stdout, _ = os.Open(os.DevNull)
	if m, err := NewMaster(masterAddr); m != nil {
		defer m.Close()
	} else if err != nil {
		t.Errorf("error creating master: %v", err)
		return
	}

	time.Sleep(1 * time.Second)

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
		t.Errorf("error posting write: %v", err)
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

	url = "http://" + masterAddr + "/read"
	encoding = "application/json"

	payload = map[string]string{
		"key": "toto",
	}
	jsonPayload, _ = json.Marshal(payload)
	buffer = bytes.NewBuffer(jsonPayload)

	readResp, err := http.Post(url, encoding, buffer)
	if err != nil {
		t.Errorf("error posting read: %v", err)
		return
	}
	defer readResp.Body.Close()
	buf.ReadFrom(readResp.Body)
	body = buf.String()

	if resp.StatusCode != 200 {
		t.Errorf("/read query failed: %v", body)
		return
	}

	if body != "le sang" {
		t.Errorf("expected \"le sang\", got \"%s\"", body)
	}
}
