package main

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

func main() {
	qb := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	builders := map[string]ToSqler{
		"select all":                       qb.Select("*").From("actor"),
		"select some":                      qb.Select("first_name", "last_name").From("actor"),
		"select filter":                    qb.Select("title").From("file").Where("rental_rate = ?", 0.99),
		"select distinct":                  qb.Select().Distinct().Column("rating").From("film"),
		"select order by & limit & offset": qb.Select("*").From("payment").OrderBy("payment_date desc").Limit(10).Offset(10),
		"select count":                     qb.Select("count(*)").From("inventory"),
		"select inner join":                qb.Select("c.first_name", "c.last_name", "a.phone").From("customer c").InnerJoin("address a on c.address_id = a.address_id"),
		"select group by":                  qb.Select("rating", "count(*) as film_count").From("film").GroupBy("rating").Having("count(*) > 200"),
		"select between":                   qb.Select("*").From("film").Where("length >= 60").Where("length <= 120"),
		"select multiple joins":            qb.Select("f.title", "c.name").From("film f").InnerJoin("film_category fc on f.film_id = fc.film_id").InnerJoin("category c on fc.category_id = c.category_id"),
		"select like":                      qb.Select("*").From("film").Where("title like ?", "W%"),
	}

	for name, builder := range builders {
		fmt.Printf("%s: ", name)
		fmt.Println(builder.ToSql())
	}
}

type ToSqler interface {
	ToSql() (string, []any, error)
}
