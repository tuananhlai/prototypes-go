package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"
)

const (
	connStr = "postgres://postgres:postgres@localhost:5432/prototype?sslmode=disable"
	addr    = ":8080"
)

func main() {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("error opening database: %v", err)
	}
	defer db.Close()

	err = setupDatabase(db)
	if err != nil {
		log.Fatalf("error setting up database: %v", err)
	}

	kvStore := NewKVStore(db)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /kv/{key}", func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")

		val, err := kvStore.Get(r.Context(), key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Add("content-type", "application/json")
		json.NewEncoder(w).Encode(GetKeyResponseDTO{
			Key:   key,
			Value: val,
		})
	})

	mux.HandleFunc("PUT /kv/{key}", func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")

		var req PutKeyRequestDTO
		if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if req.Value == "" {
			http.Error(w, "value cannot be empty", http.StatusBadRequest)
		}

		expiration, err := time.ParseDuration(req.Expiration)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		err = kvStore.Put(r.Context(), key, req.Value, expiration)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("DELETE /kv/{key}", func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")

		err := kvStore.Del(r.Context(), key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	log.Printf("kv store listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func setupDatabase(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS kv (
		key VARCHAR(255) PRIMARY KEY,
		value VARCHAR(255),
		expires_at TIMESTAMPTZ
	)`)
	if err != nil {
		return fmt.Errorf("error creating table: %v", err)
	}

	return nil
}

type KVStore struct {
	db *sql.DB
}

func NewKVStore(db *sql.DB) *KVStore {
	return &KVStore{db: db}
}

func (k *KVStore) Put(ctx context.Context, key string, value string, expiration time.Duration) error {
	expiresAt := new(time.Time)
	if expiration > 0 {
		*expiresAt = time.Now().Add(expiration)
	} else {
		expiresAt = nil
	}

	_, err := k.db.ExecContext(ctx, `
	INSERT INTO kv (key, value, expires_at) VALUES ($1, $2, $3)
	ON CONFLICT (key) DO UPDATE SET value = $2, expires_at = $3
	`, key, value, expiresAt)
	if err != nil {
		return fmt.Errorf("error inserting kv: %v", err)
	}

	return nil
}

func (k *KVStore) Get(ctx context.Context, key string) (string, error) {
	var value string

	err := k.db.QueryRowContext(ctx, `
	SELECT value FROM kv WHERE key = $1 AND (expires_at > NOW() OR expires_at IS NULL)
	`, key).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("key not found")
		}
		return "", fmt.Errorf("error scanning value: %v", err)
	}
	return value, nil
}

func (k *KVStore) Del(ctx context.Context, key string) error {
	_, err := k.db.ExecContext(ctx, `
	UPDATE kv SET expires_at = NOW() WHERE key = $1
	`, key)
	if err != nil {
		return fmt.Errorf("error deleting kv: %v", err)
	}

	return nil
}

type PutKeyRequestDTO struct {
	Value string `json:"value"`
	// Duration format string (i.e 15m)
	Expiration string `json:"expiration"`
}

type GetKeyResponseDTO struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
