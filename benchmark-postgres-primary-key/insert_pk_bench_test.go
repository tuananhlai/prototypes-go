package bench

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand/v2"
	"sync"
	"testing"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

const (
	intTable  = "bench_int_pk"
	uuidTable = "bench_uuid_pk"
	rowCount  = 1_000_000
	dsn       = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
)

var (
	db   *sql.DB
	once sync.Once
)

func getDB(tb testing.TB) *sql.DB {
	tb.Helper()

	once.Do(func() {
		var err error
		db, err = sql.Open("postgres", dsn)
		if err != nil {
			panic(err)
		}

		// Basic pool sizing; tune as needed.
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(10)
		db.SetConnMaxLifetime(30 * time.Minute)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			panic(err)
		}
	})

	return db
}

func setupSchema(tb testing.TB) {
	tb.Helper()
	d := getDB(tb)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// pgcrypto provides gen_random_uuid()
	stmts := []string{
		fmt.Sprintf(`DROP TABLE IF EXISTS %s;`, intTable),
		fmt.Sprintf(`DROP TABLE IF EXISTS %s;`, uuidTable),

		fmt.Sprintf(`
			CREATE TABLE %s (
				id  BIGSERIAL PRIMARY KEY,
				age INTEGER NOT NULL
			);`, intTable),

		fmt.Sprintf(`
			CREATE TABLE %s (
				id  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				age INTEGER NOT NULL
			);`, uuidTable),

		fmt.Sprintf(`CREATE INDEX %s_age_idx ON %s (age);`, intTable, intTable),
		fmt.Sprintf(`CREATE INDEX %s_age_idx ON %s (age);`, uuidTable, uuidTable),
	}

	for _, s := range stmts {
		if _, err := d.ExecContext(ctx, s); err != nil {
			tb.Fatalf("setup failed on %q: %v", s, err)
		}
	}
}

func truncateTable(tb testing.TB, table string, restartIdentity bool) {
	tb.Helper()
	d := getDB(tb)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	q := fmt.Sprintf("TRUNCATE TABLE %s", table)
	if restartIdentity {
		q += " RESTART IDENTITY"
	}
	q += ";"

	if _, err := d.ExecContext(ctx, q); err != nil {
		tb.Fatalf("truncate %s failed: %v", table, err)
	}
}

func copyInsertAges(tb testing.TB, table string, n int) {
	tb.Helper()
	d := getDB(tb)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	tx, err := d.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		tb.Fatalf("begin tx failed: %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Insert only "age". For int PK, id comes from sequence; for uuid PK, id comes from default gen_random_uuid().
	stmt, err := tx.PrepareContext(ctx, pq.CopyIn(table, "age"))
	if err != nil {
		tb.Fatalf("prepare copyin failed: %v", err)
	}

	rnd := rand.New(rand.NewChaCha8([32]byte{}))

	// Stream rows; no large allocations.
	for i := range n {
		age := rnd.IntN(200)
		if _, err := stmt.ExecContext(ctx, age); err != nil {
			_ = stmt.Close()
			tb.Fatalf("copy exec failed at row %d: %v", i, err)
		}
	}

	// Flush COPY
	if _, err := stmt.ExecContext(ctx); err != nil {
		_ = stmt.Close()
		tb.Fatalf("copy flush failed: %v", err)
	}
	if err := stmt.Close(); err != nil {
		tb.Fatalf("copy close failed: %v", err)
	}

	if err := tx.Commit(); err != nil {
		tb.Fatalf("commit failed: %v", err)
	}
}

func BenchmarkInsert1M_IntPK(b *testing.B) {
	setupSchema(b)

	for b.Loop() {
		b.StopTimer()
		truncateTable(b, intTable, true)
		b.StartTimer()

		copyInsertAges(b, intTable, rowCount)
	}
}

func BenchmarkInsert1M_UUIDPK(b *testing.B) {
	setupSchema(b)

	for b.Loop() {
		b.StopTimer()
		truncateTable(b, uuidTable, false)
		b.StartTimer()

		copyInsertAges(b, uuidTable, rowCount)
	}
}
