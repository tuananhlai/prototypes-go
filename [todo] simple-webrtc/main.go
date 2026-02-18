package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const addr = ":8080"

type signalMessage struct {
	ID      int64  `json:"id"`
	Room    string `json:"room"`
	From    string `json:"from"`
	To      string `json:"to"`
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

type signalStore struct {
	mu       sync.Mutex
	nextID   int64
	messages []signalMessage
}

func (s *signalStore) add(msg signalMessage) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	msg.ID = s.nextID
	s.messages = append(s.messages, msg)
	return msg.ID
}

func (s *signalStore) list(room, to string, afterID int64) []signalMessage {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make([]signalMessage, 0)
	for _, msg := range s.messages {
		if msg.Room != room || msg.To != to || msg.ID <= afterID {
			continue
		}
		result = append(result, msg)
	}
	return result
}

func main() {
	store := &signalStore{}
	mux := http.NewServeMux()

	mux.HandleFunc("/signal", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		switch r.Method {
		case http.MethodPost:
			var msg signalMessage
			if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
				http.Error(w, "invalid body", http.StatusBadRequest)
				return
			}
			if msg.Room == "" || msg.From == "" || msg.To == "" || msg.Type == "" || msg.Payload == "" {
				http.Error(w, "missing fields", http.StatusBadRequest)
				return
			}
			id := store.add(msg)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "id": id})
		case http.MethodGet:
			room := r.URL.Query().Get("room")
			to := r.URL.Query().Get("to")
			after, _ := strconv.ParseInt(r.URL.Query().Get("after"), 10, 64)
			if room == "" || to == "" {
				http.Error(w, "room and to are required", http.StatusBadRequest)
				return
			}
			messages := store.list(room, to, after)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(messages)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.Handle("/", http.FileServer(http.Dir(".")))

	s := &http.Server{
		Addr:              addr,
		Handler:           logRequests(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("Open http://localhost%s in two tabs", addr)
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
