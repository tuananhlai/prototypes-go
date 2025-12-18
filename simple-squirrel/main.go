package main

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

func main() {
	qb := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	builders := map[string]ToSqler{
		"select all":      qb.Select("*").From("actor"),
		"select some":     qb.Select("first_name", "last_name").From("actor"),
		"select filter":   qb.Select("title").From("file").Where("rental_rate = ?", 0.99),
		"select distinct": qb.Select("distinct rating").From("film"),
	}

	for name, builder := range builders {
		fmt.Printf("%s: ", name)
		fmt.Println(builder.ToSql())
	}
}

type ToSqler interface {
	ToSql() (string, []any, error)
}
