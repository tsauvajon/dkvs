package dkvs

import (
	"context"
	"encoding/json"
	"fmt"
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

	h.HandleFunc("/write", t.writeHandler)
	h.HandleFunc("/read", t.readHandler)
	h.HandleFunc("/list", t.listHandler)
	h.HandleFunc("/join", t.joinHandler)

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
	fmt.Fprint(w, string(val))
}

func (t *httpTransport) listHandler(w http.ResponseWriter, r *http.Request) {
	val, err := t.List()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, val)
}

func (t *httpTransport) joinHandler(w http.ResponseWriter, r *http.Request) {
	var p struct {
		Node *Node `json:"node"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}

	err := t.Join(p.Node)
	if err != nil {
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

func (t *httpTransport) List() ([]*Node, error) {
	return t.n.ListNodes()

}

func (t *httpTransport) Join(slave *Node) error {
	return t.n.JoinMaster(slave)
}
