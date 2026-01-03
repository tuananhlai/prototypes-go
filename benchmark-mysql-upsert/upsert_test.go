package benchmarkmysqlupsert_test

import (
	"database/sql"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

const (
	// multiStatements is used only for init() schema setup.
	connStr = "root:root@tcp(localhost:3306)/prototype?multiStatements=true&parseTime=true"
)

var (
	db *sql.DB
)

func init() {
	var err error
	db, err = sql.Open("mysql", connStr)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`
DROP TABLE IF EXISTS users;
CREATE TABLE users (
	id  INT PRIMARY KEY,
	age INT NOT NULL
);
-- without this index, two benchmarks will perform similarly.
ALTER TABLE users ADD INDEX idx_age (age);
INSERT INTO users (id, age) VALUES (1, 25);
`)
	if err != nil {
		panic(err)
	}
}

func BenchmarkOnDuplicateKeyUpdate(b *testing.B) {
	// Prepare once; measure execution only.
	tx, err := db.Begin()
	if err != nil {
		b.Fatal(err)
	}
	defer tx.Commit()

	stmt, err := tx.Prepare(`
INSERT INTO users (id, age)
VALUES (?, ?)
ON DUPLICATE KEY UPDATE age = VALUES(age)
`)
	if err != nil {
		b.Fatal(err)
	}
	defer stmt.Close()

	age := 25
	for b.Loop() {
		age ^= 1 // alternate 25/24 to force an actual update on duplicates
		_, err := stmt.Exec(1, age)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReplaceInto(b *testing.B) {
	tx, err := db.Begin()
	if err != nil {
		b.Fatal(err)
	}
	defer tx.Commit()

	stmt, err := tx.Prepare(`REPLACE INTO users (id, age) VALUES (?, ?)`)
	if err != nil {
		b.Fatal(err)
	}
	defer stmt.Close()

	age := 50
	for b.Loop() {
		age ^= 1 // alternate values so the row content changes
		_, err := stmt.Exec(1, age)
		if err != nil {
			b.Fatal(err)
		}
	}
}
