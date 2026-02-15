package main

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Drop struct {
	Payload   []byte
	ExpiresAt time.Time
}

type Node struct {
	Peers []string
	Store map[string]Drop
	Mu    sync.Mutex
}

func NewNode(peers []string) *Node {
	return &Node{
		Peers: peers,
		Store: make(map[string]Drop),
	}
}

type DropRequest struct {
	ID      string `json:"id"`
	Payload string `json:"payload"`
	TTL     int    `json:"ttl"`
}

func (n *Node) DropHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DropRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	data, err := base64.StdEncoding.DecodeString(req.Payload)
	if err != nil {
		http.Error(w, "bad payload", http.StatusBadRequest)
		return
	}

	n.Mu.Lock()
	n.Store[req.ID] = Drop{
		Payload:   data,
		ExpiresAt: time.Now().Add(time.Duration(req.TTL) * time.Second),
	}
	n.Mu.Unlock()

	w.WriteHeader(http.StatusCreated)
}

func (n *Node) ReceiveHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/receive/")

	n.Mu.Lock()
	drop, ok := n.Store[id]
	if ok {
		delete(n.Store, id)
	}
	n.Mu.Unlock()

	if ok && time.Now().Before(drop.ExpiresAt) {
		resp := map[string]string{
			"payload": base64.StdEncoding.EncodeToString(drop.Payload),
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	for _, peer := range n.Peers {
		resp, err := http.Get(peer + "/receive/" + id)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			w.WriteHeader(http.StatusOK)
			io.Copy(w, resp.Body)
			return
		}
	}

	http.NotFound(w, r)
}

func (n *Node) CleanupLoop() {
	for {
		time.Sleep(10 * time.Second)
		now := time.Now()

		n.Mu.Lock()
		for id, drop := range n.Store {
			if now.After(drop.ExpiresAt) {
				delete(n.Store, id)
			}
		}
		n.Mu.Unlock()
	}
}

func main() {
	node := NewNode([]string{
		"http://localhost:8081",
		"http://localhost:8082",
	})

	go node.CleanupLoop()

	http.HandleFunc("/drop", node.DropHandler)
	http.HandleFunc("/receive/", node.ReceiveHandler)

	log.Println("Node running at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
