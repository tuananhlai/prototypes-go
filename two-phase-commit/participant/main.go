package main

import (
	"encoding/json"
	"net/http"
	"sync"
)

func main() {

}

type TxState string

const (
	StateInit      TxState = "INIT"
	StatePrepared  TxState = "PREPARED"
	StateCommitted TxState = "COMMITTED"
	StateAborted   TxState = "ABORTED"
)

type PrepareRequest struct {
	TxID   string            `json:"tx_id"`
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
			_ = json.NewEncoder(w).Encode(PrepareResponse{OK: true, Message: "already prepared"})
			return
		case StateCommitted:
			_ = json.NewEncoder(w).Encode(PrepareResponse{OK: true, Message: "already commited"})
			return
		case StateAborted:
			_ = json.NewEncoder(w).Encode(PrepareResponse{OK: false, Message: "already aborted"})
			return
		}
	}

	if !p.allowPrepare {
		p.txState[req.TxID] = StateAborted
		_ = json.NewEncoder(w).Encode(PrepareResponse{OK: false, Message: "prepare rejected"})
	}
}
