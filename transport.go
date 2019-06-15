package dkvs

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"
)

// Transport describes the communication layer
type Transport interface {
	Start(addr string) error
	Stop() error

	Write(key, val string) error
	Read(key string) (string, error)
	List() ([]*Node, error)
}

// NewHTTPTransport creates an http transport
func NewHTTPTransport() Transport {
	return &httpTransport{}
}

type httpTransport struct {
	srv *http.Server
}

func (t *httpTransport) Start(addr string) error {
	h := http.NewServeMux()

	h.HandleFunc("/write", t.writeHandler)
	h.HandleFunc("/read", t.readHandler)
	h.HandleFunc("/list", t.listHandler)

	t.srv = &http.Server{Addr: addr, Handler: h}

	go func() {
		if err := t.srv.ListenAndServe(); err != nil {
			fmt.Print(err)
		}
	}()

	// Setting up signal capturing
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

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
	err := t.Write("", "")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func (t *httpTransport) readHandler(w http.ResponseWriter, r *http.Request) {
	val, err := t.Read("")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, val)
	}
}

func (t *httpTransport) listHandler(w http.ResponseWriter, r *http.Request) {
	val, err := t.List()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, val)
	}
}

func (t *httpTransport) Write(key, val string) error {
	return errorNotImplemented
}

func (t *httpTransport) Read(key string) (string, error) {
	return "", errorNotImplemented
}

func (t *httpTransport) List() ([]*Node, error) {
	return nil, errorNotImplemented
}
