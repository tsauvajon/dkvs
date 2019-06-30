package dkvs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Transport describes the communication layer
type Transport interface {
	Start(n *Node) error
	Stop() error

	Write(key, val string) error
	Read(key string) ([]byte, error)
	List() ([]*Node, error)

	Join(slave *Node) error
}

// NewHTTPTransport creates an http transport
func NewHTTPTransport() Transport {
	return &httpTransport{}
}

type httpTransport struct {
	srv *http.Server
	n   *Node
}

func (t *httpTransport) Start(n *Node) error {
	t.n = n
	h := http.NewServeMux()

	// simplistic routes; using gRPC or a REST API would be better practice
	h.HandleFunc("/write", t.writeHandler)
	h.HandleFunc("/read", t.readHandler)
	h.HandleFunc("/multi", t.multiHandler)
	h.HandleFunc("/list", t.listHandler)
	h.HandleFunc("/join", t.joinHandler)
	h.HandleFunc("/update", t.updateHandler)
	h.HandleFunc("/receive", t.receiveHandler)
	h.HandleFunc("/replicate", t.replicateHandler)

	t.srv = &http.Server{Addr: t.n.Address, Handler: h}

	go func() {
		if err := t.srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	// Setting up signal capturing
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	signal.Notify(stop, syscall.SIGTERM)
	signal.Notify(stop, syscall.SIGKILL)

	// Waiting for SIGINT (pkill -2)
	<-stop

	// Shutdown gracefully or after 1 sec
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := t.srv.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}

func (t *httpTransport) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	return t.srv.Shutdown(ctx)
}

func (t *httpTransport) writeHandler(w http.ResponseWriter, r *http.Request) {
	var p struct {
		Key   string `json:"key"`
		Value string `json:"val"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}

	err := t.Write(p.Key, p.Value)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (t *httpTransport) receiveHandler(w http.ResponseWriter, r *http.Request) {
	var p struct {
		Key   string `json:"key"`
		Value string `json:"val"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}

	err := t.Receive(p.Key, p.Value)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (t *httpTransport) readHandler(w http.ResponseWriter, r *http.Request) {
	var p struct {
		Key string `json:"key"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}

	val, err := t.Read(p.Key)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(val)
}

func (t *httpTransport) multiHandler(w http.ResponseWriter, r *http.Request) {
	var p struct {
		Keys []string `json:"keys"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}

	values, err := t.Multi(p.Keys)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(values)
}

func (t *httpTransport) listHandler(w http.ResponseWriter, r *http.Request) {
	val, err := t.List()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)

	jsonVal, err := json.Marshal(val)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return
	}

	w.Write(jsonVal)
}

func (t *httpTransport) joinHandler(w http.ResponseWriter, r *http.Request) {
	var p *Node

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}

	err := t.Join(p)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (t *httpTransport) updateHandler(w http.ResponseWriter, r *http.Request) {
	var p map[string]*Node

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}

	err := t.Update(p)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (t *httpTransport) replicateHandler(w http.ResponseWriter, r *http.Request) {
	if err := t.Replicate(r.Body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (t *httpTransport) Write(key, val string) error {
	return t.n.WriteValue(key, val)
}

func (t *httpTransport) Read(key string) ([]byte, error) {
	return t.n.ReadValue(key)

}

func (t *httpTransport) Multi(keys []string) ([]byte, error) {
	return t.n.ReadMultipleValues(keys...)

}

func (t *httpTransport) List() ([]*Node, error) {
	return t.n.ListNodes()

}

func (t *httpTransport) Join(slave *Node) error {
	return t.n.Join(slave)
}

func (t *httpTransport) Update(nodes map[string]*Node) error {
	return t.n.ReceiveListUpdate(nodes)
}

func (t *httpTransport) Receive(key, val string) error {
	return t.n.ReceiveWrite(key, val)
}

func (t *httpTransport) Replicate(r io.Reader) error {
	return t.n.ReplicateFromMaster(r)
}
