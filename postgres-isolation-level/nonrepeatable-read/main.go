package main

import (
	"context"
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

const (
	dsn = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
)

func main() {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}

	err = setupDB(db)
	if err != nil {
		log.Fatal(err)
	}

	// demoReadUncommited(db)
	// demoReadCommitted(db)
	demoRepeatableRead(db)
}

func setupDB(db *sql.DB) error {
	_, err := db.Exec(`
drop table if exists accounts;
create table accounts (
	id int primary key,
	balance int
);
insert into accounts (id, balance) values (1, 100);
		`)

	return err
}

// While Postgres accepts read uncommitted as an isolation level, it isn't actually supported.
// https://www.postgresql.org/docs/current/transaction-iso.html#:~:text=PostgreSQL%27s%20Read%20Uncommitted%20mode%20behaves%20like%20Read%20Committed
func demoReadUncommited(db *sql.DB) {
	globalCtx := context.Background()
	tx1, err := db.BeginTx(globalCtx, &sql.TxOptions{
		Isolation: sql.LevelReadUncommitted,
	})
	assertNoError(err)
	defer tx1.Rollback()

	balance, err := getBalance(tx1)
	assertNoError(err)
	log.Println("Tx 1 reads:", balance)

	tx2, err := db.BeginTx(globalCtx, &sql.TxOptions{
		Isolation: sql.LevelReadUncommitted,
	})
	assertNoError(err)
	defer tx2.Rollback()

	err = updateBalance(tx2, 150)
	assertNoError(err)
	log.Println("Tx 2 updates balance to 150")

	balance, err = getBalance(tx1)
	assertNoError(err)
	log.Println("Tx 1 reads again:", balance)
}

func demoReadCommitted(db *sql.DB) {
	globalCtx := context.Background()
	tx1, err := db.BeginTx(globalCtx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	assertNoError(err)
	defer tx1.Rollback()

	balance, err := getBalance(tx1)
	assertNoError(err)
	log.Println("Tx 1 reads:", balance)

	err = updateBalance(db, 150)
	assertNoError(err)
	log.Println("Tx 2 updates balance to 150")

	balance, err = getBalance(tx1)
	assertNoError(err)
	log.Println("Tx 1 reads again:", balance)
}

func demoRepeatableRead(db *sql.DB) {
	globalCtx := context.Background()
	tx1, err := db.BeginTx(globalCtx, &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
	})
	assertNoError(err)
	defer tx1.Rollback()

	balance, err := getBalance(tx1)
	assertNoError(err)
	log.Println("Tx 1 reads:", balance)

	err = updateBalance(db, 150)
	assertNoError(err)
	log.Println("Tx 2 updates balance to 150")

	balance, err = getBalance(tx1)
	assertNoError(err)
	log.Println("Tx 1 reads again:", balance)
}

func assertNoError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func getBalance(tx *sql.Tx) (int, error) {
	var balance int
	err := tx.QueryRow("select balance from accounts where id = 1").Scan(&balance)
	if err != nil {
		return 0, err
	}
	return balance, nil
}

func updateBalance(execer Execer, newBalance int) error {
	_, err := execer.Exec("update accounts set balance = $1 where id = 1", newBalance)
	if err != nil {
		return err
	}
	return nil
}

type Execer interface {
	Exec(query string, args ...any) (sql.Result, error)
}
