package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type PrepareRequest struct {
	TxID   string            `json:"tx_id"`
	Writes map[string]string `json:"writes"`
}

type PrepareResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
}

type DecisionRequest struct {
	TxID string `json:"tx_id"`
}

type Coordinator struct {
	client       *http.Client
	participants []string
	timeout      time.Duration
}

func NewCoordinator(participants []string, timeout time.Duration) *Coordinator {
	return &Coordinator{
		client: &http.Client{
			Timeout: timeout,
		},
		participants: participants,
		timeout:      timeout,
	}
}

// postJSON encodes the request body into JSON and sends it to the given endpoint using POST request. If
// available, the response will be unmarshaled into `out`.
func postJSON(ctx context.Context, c *http.Client, url string, v any, out any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, url, "application/json", bytes.NewReader(b))
	if err != nil {
		return err
	}

	res, err := c.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 && res.StatusCode < 600 {
		return fmt.Errorf("http %d", res.StatusCode)
	}

	if out != nil {
		return json.NewDecoder(res.Body).Decode(out)
	}

	return nil
}

func (co *Coordinator) Run2PC(txID string, writesPerParticipant []map[string]string) (decision string, err error) {
	if len(writesPerParticipant) != len(co.participants) {
		return "", fmt.Errorf("writesPerParticipant must match participants length")
	}

	log.Printf("[coord] tx=$s phase=PREPARE\n", txID)
	allOK := true

	for i, base := range co.participants {
		ctx, cancel := context.WithTimeout(context.Background(), co.timeout)
		defer cancel()

		var res PrepareResponse
		req := PrepareRequest{TxID: txID, Writes: writesPerParticipant[i]}
		err := postJSON(ctx, co.client, base+"/prepare", req, &res)
		if err != nil || !res.OK {
			allOK = false
			log.Printf("[coord] tx=%s participant=%s prepare=FAIL err=%v resp=%+v\n", txID, base, err, res)
		} else {
			log.Printf("[coord] tx=%s participant=%s prepare=OK\n", txID, base)
		}
	}

	if allOK {
		decision = "COMMIT"
	} else {
		decision = "ABORT"
	}

	log.Printf("[coord] tx=%s decision=%s phase=DECIDE\n", txID, decision)

	for _, base := range co.participants {

	}
}
