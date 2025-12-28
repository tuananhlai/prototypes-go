package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"
)

const (
	connStr = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	addr    = ":8080"
)

// TODO: prevent database already exists error when running the example multiple times.
// TODO: add logging so that reader knows which shard is being used for each request.
// TODO: support multiple sharding strategies.
// TODO: reuse hash.New32a() to avoid creating a new hash function for each request.
func main() {
	shards, err := setupDatabase(connStr)
	if err != nil {
		log.Fatalf("error setting up database: %v", err)
	}

	shardManager, err := NewShardManager(shards)
	if err != nil {
		log.Fatalf("error creating shard manager: %v", err)
	}

	kvStore := NewKVStore(shardManager)

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

// setupDatabase creates and initializes 3 database shards.
func setupDatabase(connStr string) ([]*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	shardNames := []string{"kvstore_1", "kvstore_2", "kvstore_3"}
	shards := []*sql.DB{}

	var errs []error
	for _, shardName := range shardNames {
		_, err := db.Exec(fmt.Sprintf("CREATE DATABASE %s", shardName))
		if err != nil {
			errs = append(errs, err)
			continue
		}

		shard, err := sql.Open("postgres", fmt.Sprintf(
			"postgres://postgres:postgres@localhost:5432/%s?sslmode=disable", shardName))
		if err != nil {
			errs = append(errs, err)
			continue
		}
		shards = append(shards, shard)

		_, err = shard.Exec(`
			CREATE TABLE IF NOT EXISTS kv (
				key        VARCHAR(255) PRIMARY KEY,
				value      VARCHAR(255),
				expires_at TIMESTAMPTZ
			)
		`)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("error setting up database: %v", errs)
	}

	return shards, nil
}

type KVStore struct {
	shardManager *ShardManager
}

func NewKVStore(shardManager *ShardManager) *KVStore {
	return &KVStore{shardManager: shardManager}
}

func (k *KVStore) Put(ctx context.Context, key string, value string, expiration time.Duration) error {
	expiresAt := new(time.Time)
	if expiration > 0 {
		*expiresAt = time.Now().Add(expiration)
	} else {
		expiresAt = nil
	}

	_, err := k.shardManager.GetShard(key).ExecContext(ctx, `
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

	err := k.shardManager.GetShard(key).QueryRowContext(ctx, `
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
	_, err := k.shardManager.GetShard(key).ExecContext(ctx, `
	UPDATE kv SET expires_at = NOW() WHERE key = $1
	`, key)
	if err != nil {
		return fmt.Errorf("error deleting kv: %v", err)
	}

	return nil
}

// ShardManager picks a shard for a given record key.
type ShardManager struct {
	shards []*sql.DB
}

func NewShardManager(shards []*sql.DB) (*ShardManager, error) {
	if len(shards) == 0 {
		return nil, fmt.Errorf("no data sources provided")
	}

	return &ShardManager{
		shards: shards,
	}, nil
}

// GetShard returns the shard for the given key.
func (sm *ShardManager) GetShard(key string) *sql.DB {
	hash := fnv.New32a()
	hash.Write([]byte(key))
	index := hash.Sum32() % uint32(len(sm.shards))
	return sm.shards[index]
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
