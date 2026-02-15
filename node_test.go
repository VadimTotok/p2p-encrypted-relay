package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDropAndReceive(t *testing.T) {
	node := NewNode(nil)

	payload := base64.StdEncoding.EncodeToString([]byte("hello"))
	body, _ := json.Marshal(DropRequest{
		ID:      "x1",
		Payload: payload,
		TTL:     5,
	})

	req := httptest.NewRequest("POST", "/drop", bytes.NewReader(body))
	w := httptest.NewRecorder()
	node.DropHandler(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	req2 := httptest.NewRequest("GET", "/receive/x1", nil)
	w2 := httptest.NewRecorder()
	node.ReceiveHandler(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w2.Code)
	}

	var resp map[string]string
	json.NewDecoder(w2.Body).Decode(&resp)

	decoded, _ := base64.StdEncoding.DecodeString(resp["payload"])
	if string(decoded) != "hello" {
		t.Fatalf("unexpected payload: %s", decoded)
	}
}

func TestExpiredDrop(t *testing.T) {
	node := NewNode(nil)

	node.Mu.Lock()
	node.Store["exp"] = Drop{
		Payload:   []byte("x"),
		ExpiresAt: time.Now().Add(-1 * time.Second),
	}
	node.Mu.Unlock()

	req := httptest.NewRequest("GET", "/receive/exp", nil)
	w := httptest.NewRecorder()
	node.ReceiveHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
