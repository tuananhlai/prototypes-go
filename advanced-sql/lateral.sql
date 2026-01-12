/*
LATERAL Joins in SQL

Definition:
A LATERAL join is a special type of SQL join that allows a subquery in
the FROM clause to reference columns from tables that appear earlier in
the FROM clause. The subquery on the right side of a LATERAL join can
access and use values from the row of the table(s) on the left. In
PostgreSQL and some other databases, you can write this using the
LATERAL keyword, either with CROSS JOIN LATERAL or just LATERAL in the
FROM list.

Syntax Example:
SELECT ...
FROM outer_table
CROSS JOIN LATERAL (
    SELECT ...
    FROM inner_table
    WHERE inner_table.col = outer_table.col
) AS alias

Use Cases:
1. Per-row Top-N Queries: Select the top N related rows for each row in
   another table. For example, the 3 most recent rentals for each
   customer.
2. Correlated Aggregates: Compute aggregates in a subquery that depend
   on the value of the current row in an outer table.
3. Filtering by Calculated Subqueries: Filter rows or add columns that
   require correlation with each row of an outer table.
4. Expanding Arrays or JSON: Use LATERAL to unnest arrays or JSON
   values for each outer row, where the subquery needs information from
   the outer row.
5. Complex Derived Data: Generate calculated or derived results per row
   where the calculation depends on both the outer and inner tables.

Why use LATERAL:
Without LATERAL, subqueries in the FROM clause cannot access values
from previous tables in the FROM clause. LATERAL provides row-wise
correlation and greater flexibility for advanced query constructs.

Performance Implications and Pitfalls:
- Performance: The subquery in a LATERAL join runs for every row from
  the outer table. This can be much slower than a regular join or
  non-correlated subquery, especially on large tables or when the
  subquery is complex. Try to make the LATERAL subquery as efficient as
  possible, using indexes, good filtering, and limits when needed.
- Index Utilization: LATERAL join performance relies on indexes in the
  correlated subquery. Lack of indexes can lead to repeated sequential
  scans, making the query slow.
- Scalability: Because LATERAL triggers the subquery for every outer
  row, query time can grow rapidly as data size increases. Test your
  queries on realistic data sizes before using them in production.
- Optimization Limitations: Query planners are less able to optimize a
  LATERAL subquery since it depends on outer row values. Execution
  plans may be less efficient as a result.
- Logical Pitfall: If the LATERAL subquery returns multiple rows for one
  row from the outer table, this may create more result rows than
  expected.
- General Advice: Use LATERAL when you actually need per-row
  correlation. If you don't need this, consider a regular join or
  set-based operations for better performance.
*/

-- Find the 3 most recent rentals for each customer.
SELECT
    c.customer_id,
    c.first_name,
    c.last_name,
    r.rental_id,
    r.rental_date
FROM
    customer AS c
CROSS JOIN
    LATERAL (
        SELECT
            r.rental_id,
            r.rental_date
        FROM
            rental AS r
        WHERE
            r.customer_id = c.customer_id
        ORDER BY
            r.rental_date DESC
        LIMIT
            3
    ) AS r
ORDER BY
    c.customer_id ASC,
    r.rental_date DESC;

-- Find the most rented film for each customer.
SELECT
    c.customer_id,
    c.first_name,
    c.last_name,
    f.title,
    x.rent_count
FROM
    customer AS c
CROSS JOIN
    LATERAL (
        SELECT
            i.film_id,
            count(*) AS rent_count
        FROM
            rental AS r
        INNER JOIN inventory AS i ON r.inventory_id = i.inventory_id
        WHERE
            r.customer_id = c.customer_id
        GROUP BY
            i.film_id
        ORDER BY
            rent_count DESC
        LIMIT
            1
    ) AS x
INNER JOIN film AS f ON x.film_id = f.film_id;

-- Latest rental per store
SELECT
    s.store_id,
    r.rental_id,
    r.rental_date
FROM
    store AS s
LEFT JOIN LATERAL (
    SELECT
        r.rental_id,
        r.rental_date
    FROM
        rental AS r
    INNER JOIN inventory AS i ON r.inventory_id = i.inventory_id
    WHERE
        i.store_id = s.store_id
    ORDER BY
        r.rental_date DESC
    LIMIT
        1
) AS r ON TRUE
ORDER BY
    r.rental_date DESC NULLS LAST;
