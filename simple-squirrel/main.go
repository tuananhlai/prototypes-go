package main

import (
	"log"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const (
	connStr = "postgres://postgres:postgres@localhost:25432/pagila?sslmode=disable"
)

func main() {
	db := sqlx.MustOpen("postgres", connStr)
	qb := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(db)

	var res int
	if err := qb.Select("1").QueryRow().Scan(&res); err != nil {
		log.Fatalf("error: %v", err)
	}

	selectAll(db, qb)
}

func selectAll(db *sqlx.DB, qb sq.StatementBuilderType) {
	query, _, err := qb.Select("actor_id", "first_name", "last_name", "last_update").From("actor").ToSql()
	if err != nil {
		log.Fatal(err)
	}
	rows, err := db.Queryx(query)
	if err != nil {
		log.Fatal(err)
	}

	var actors []Actor
	for rows.Next() {
		var actor Actor
		err = rows.StructScan(&actor)
		if err != nil {
			log.Fatal(err)
		}

		actors = append(actors, actor)
	}

	log.Println(len(actors), actors[:10])
}

type Actor struct {
	ActorID    int       `db:"actor_id"`
	FirstName  string    `db:"first_name"`
	LastName   string    `db:"last_name"`
	LastUpdate time.Time `db:"last_update"`
}
