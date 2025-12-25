package main

import (
	"encoding/json"
	"fmt"
	"log"
	"maps"
	"net/http"
	"os"
	"sync"
	"time"
)

func main() {
	allow := os.Getenv("ALLOW_PREPARE") != "false"
	port := os.Getenv("PORT")
	if port == "" {
		port = "9090"
	}
	addr := fmt.Sprintf(":%s", port)

	p := NewParticipant(allow)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /prepare", p.handlePrepare)
	mux.HandleFunc("POST /commit", p.handleCommit)
	mux.HandleFunc("POST /abort", p.handleAbort)
	mux.HandleFunc("POST /dump", p.handleDump)

	log.Printf("[participant] listening on %s (ALLOW_PREPARE=%v)\n", addr, allow)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("error starting server")
	}
}

type TxState string

const (
	StateInit      TxState = "INIT"
	StatePrepared  TxState = "PREPARED"
	StateCommitted TxState = "COMMITTED"
	StateAborted   TxState = "ABORTED"
)

type PrepareRequest struct {
	TxID string `json:"tx_id"`
	// Writes are the data to be written into durable stores owned by this participant.
	Writes map[string]string `json:"writes"`
}

type PrepareResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

type DecisionRequest struct {
	TxID string `json:"tx_id"`
}

type Participant struct {
	mu sync.Mutex
	db map[string]string

	txState map[string]TxState
	pending map[string]map[string]string

	// If false, this participant will always return error on PREPARE request.
	allowPrepare bool
}

func NewParticipant(allowPrepare bool) *Participant {
	return &Participant{
		db:           make(map[string]string),
		txState:      make(map[string]TxState),
		pending:      make(map[string]map[string]string),
		allowPrepare: allowPrepare,
	}
}

func (p *Participant) handlePrepare(w http.ResponseWriter, r *http.Request) {
	var req PrepareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if st, ok := p.txState[req.TxID]; ok {
		switch st {
		case StatePrepared:
			json.NewEncoder(w).Encode(PrepareResponse{OK: true, Message: "already prepared"})
			return
		case StateCommitted:
			json.NewEncoder(w).Encode(PrepareResponse{OK: true, Message: "already commited"})
			return
		case StateAborted:
			json.NewEncoder(w).Encode(PrepareResponse{OK: false, Message: "already aborted"})
			return
		}
	}

	if !p.allowPrepare {
		p.txState[req.TxID] = StateAborted
		json.NewEncoder(w).Encode(PrepareResponse{OK: false, Message: "prepare rejected"})
	}

	staged := make(map[string]string, len(req.Writes))
	maps.Copy(staged, req.Writes)
	p.pending[req.TxID] = staged
	p.txState[req.TxID] = StatePrepared

	json.NewEncoder(w).Encode(PrepareResponse{OK: true, Message: "prepared"})
}

func (p *Participant) handleCommit(w http.ResponseWriter, r *http.Request) {
	var req DecisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	st := p.txState[req.TxID]
	if st == StateCommitted {
		json.NewEncoder(w).Encode(map[string]any{
			"ok":      true,
			"message": "already commited",
		})
		return
	}
	if st == StateAborted {
		json.NewEncoder(w).Encode(map[string]any{
			"ok":      false,
			"message": "already aborted",
		})
		return
	}

	if st != StatePrepared {
		p.txState[req.TxID] = StateAborted
		json.NewEncoder(w).Encode(map[string]any{
			"ok":      false,
			"message": "not prepared; aborting",
		})
		return
	}

	maps.Copy(p.db, p.pending[req.TxID])
	delete(p.pending, req.TxID)
	p.txState[req.TxID] = StateCommitted

	json.NewEncoder(w).Encode(map[string]any{
		"ok":      true,
		"message": "commited",
	})
}

func (p *Participant) handleAbort(w http.ResponseWriter, r *http.Request) {
	var req DecisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.txState[req.TxID] == StateCommitted {
		json.NewEncoder(w).Encode(map[string]any{
			"ok":      false,
			"message": "already commited",
		})
		return
	}

	delete(p.pending, req.TxID)
	p.txState[req.TxID] = StateAborted

	json.NewEncoder(w).Encode(map[string]any{
		"ok":      true,
		"message": "aborted",
	})
}

func (p *Participant) handleDump(w http.ResponseWriter, r *http.Request) {
	p.mu.Lock()
	defer p.mu.Unlock()

	json.NewEncoder(w).Encode(map[string]any{
		"db":      p.db,
		"txState": p.txState,
		"pending": p.pending,
		"time":    time.Now().Format(time.RFC3339Nano),
	})
}
