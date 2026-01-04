package bench

import (
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

const (
	connStr = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	numRows = 1000000
)

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`
DROP TABLE IF EXISTS bench_int_pk;
CREATE TABLE bench_int_pk (
		id BIGSERIAL PRIMARY KEY,
		age INT NOT NULL
);

DROP TABLE IF EXISTS bench_uuid_pk;
CREATE TABLE bench_uuid_pk (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		age INT NOT NULL
);
`)
	if err != nil {
		panic(err)
	}
}

func BenchmarkInsertIntPk(b *testing.B) {
	var err error
	for b.Loop() {
		err = truncateTable("bench_int_pk")
		if err != nil {
			b.Fatal(err)
		}

		err = insertIntPk(numRows)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkInsertUUIDv4PkDBGen(b *testing.B) {
	var err error
	for b.Loop() {
		err = truncateTable("bench_uuid_pk")
		if err != nil {
			b.Fatal(err)
		}

		err = insertUUIDv4PkDBGen(numRows)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkInsertUUIDv4PkAppGen(b *testing.B) {
	var err error
	for b.Loop() {
		err = truncateTable("bench_uuid_pk")
		if err != nil {
			b.Fatal(err)
		}

		err = insertUUIDv4PkAppGen(numRows)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkInsertUUIDv7PkAppGen(b *testing.B) {
	var err error
	for b.Loop() {
		err = truncateTable("bench_uuid_pk")
		if err != nil {
			b.Fatal(err)
		}

		err = insertUUIDv7PkAppGen(numRows)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func truncateTable(tableName string) error {
	_, err := db.Exec(`TRUNCATE TABLE ` + tableName)
	if err != nil {
		return err
	}
	return nil
}

// insertIntPk inserts numRows into the bench_int_pk table, which uses `BIGSERIAL` as the primary key.
func insertIntPk(numRows int) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Use COPY command for better performance when inserting a large number of rows.
	// https://stackoverflow.com/questions/46715354/how-does-copy-work-and-why-is-it-so-much-faster-than-insert
	stmt, err := tx.Prepare(pq.CopyIn("bench_int_pk", "age"))
	if err != nil {
		return err
	}

	for i := range numRows {
		_, err := stmt.Exec(i)
		if err != nil {
			return err
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		return err
	}

	err = stmt.Close()
	if err != nil {
		return err
	}

	return tx.Commit()
}

// insertUUIDv4PkDBGen inserts numRows into the bench_uuid_pk table using the database generated UUIDv4.
func insertUUIDv4PkDBGen(numRows int) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(pq.CopyIn("bench_uuid_pk", "age"))
	if err != nil {
		return err
	}

	for i := range numRows {
		_, err := stmt.Exec(i)
		if err != nil {
			return err
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		return err
	}

	err = stmt.Close()
	if err != nil {
		return err
	}

	return tx.Commit()
}

// insertUUIDv4PkAppGen inserts numRows into the bench_uuid_pk table using the application generated UUIDv4.
func insertUUIDv4PkAppGen(numRows int) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(pq.CopyIn("bench_uuid_pk", "id", "age"))
	if err != nil {
		return err
	}

	for i := range numRows {
		id, _ := uuid.NewRandom()
		_, err := stmt.Exec(id, i)
		if err != nil {
			return err
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		return err
	}

	err = stmt.Close()
	if err != nil {
		return err
	}

	return tx.Commit()
}

// insertUUIDv7PkAppGen inserts numRows into the bench_uuid_pk table using the application generated UUIDv7.
func insertUUIDv7PkAppGen(numRows int) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(pq.CopyIn("bench_uuid_pk", "id", "age"))
	if err != nil {
		return err
	}

	for i := range numRows {
		id, _ := uuid.NewV7()
		_, err = stmt.Exec(id, i)
		if err != nil {
			return err
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		return err
	}

	err = stmt.Close()
	if err != nil {
		return err
	}

	return tx.Commit()
}
