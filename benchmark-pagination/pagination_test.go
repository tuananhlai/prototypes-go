package benchmarkpagination_test

import (
	"database/sql"
	"testing"

	"github.com/go-faker/faker/v4"
	_ "github.com/lib/pq"
)

const (
	dsn = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
)

var (
	db *sql.DB
)

func init() {
	var err error
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`
DROP TABLE IF EXISTS users;
CREATE TABLE users (
	id BIGSERIAL PRIMARY KEY,
	display_name VARCHAR(255) NOT NULL,
	email VARCHAR(255) NOT NULL,
	is_active BOOLEAN NOT NULL DEFAULT FALSE
);
`)
	if err != nil {
		panic(err)
	}

	err = seedDatabase(db)
	if err != nil {
		panic(err)
	}
}

// Seed 100000 rows into `users` database using github.com/go-faker/faker/v4.
func seedDatabase(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`
INSERT INTO users (display_name, email, is_active)
VALUES ($1, $2, $3);
`)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	defer stmt.Close()

	for i := range 100000 {
		displayName := faker.Name()
		email := faker.Email()
		isActive := i%2 == 0
		if _, err := stmt.Exec(displayName, email, isActive); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func BenchmarkLimitOffset0(b *testing.B) {
	benchmarkLimitOffset(b, 0)
}

func BenchmarkLimitOffset10000(b *testing.B) {
	benchmarkLimitOffset(b, 10000)
}

func BenchmarkLimitOffset90000(b *testing.B) {
	benchmarkLimitOffset(b, 90000)
}

func BenchmarkCursor0(b *testing.B) {
	benchmarkCursor(b, 0)
}


func BenchmarkCursor10000(b *testing.B) {
	benchmarkCursor(b, 10000)
}

func BenchmarkCursor90000(b *testing.B) {
	benchmarkCursor(b, 90000)
}

func benchmarkLimitOffset(b *testing.B, offset int) {
	for b.Loop() {
		rows, err := db.Query(`
SELECT id, display_name, email, is_active
FROM users
ORDER BY id
LIMIT 10
OFFSET $1;
`, offset)
		if err != nil {
			b.Fatalf("query limit/offset: %v", err)
		}
		for rows.Next() {
			var id int64
			var displayName, email string
			var isActive bool
			if err := rows.Scan(&id, &displayName, &email, &isActive); err != nil {
				_ = rows.Close()
				b.Fatalf("scan: %v", err)
			}
		}
		if err := rows.Err(); err != nil {
			_ = rows.Close()
			b.Fatalf("rows: %v", err)
		}
		if err := rows.Close(); err != nil {
			b.Fatalf("close rows: %v", err)
		}
	}
}

func benchmarkCursor(b *testing.B, cursor int) {
	for b.Loop() {
		rows, err := db.Query(`
SELECT id, display_name, email, is_active
FROM users
WHERE id > $1
ORDER BY id
LIMIT 10;
`, cursor)
		if err != nil {
			b.Fatalf("query cursor: %v", err)
		}
		for rows.Next() {
			var id int64
			var displayName, email string
			var isActive bool
			if err := rows.Scan(&id, &displayName, &email, &isActive); err != nil {
				_ = rows.Close()
				b.Fatalf("scan: %v", err)
			}
		}
		if err := rows.Err(); err != nil {
			_ = rows.Close()
			b.Fatalf("rows: %v", err)
		}
		if err := rows.Close(); err != nil {
			b.Fatalf("close rows: %v", err)
		}
	}
}
