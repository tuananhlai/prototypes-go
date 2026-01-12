/*
Window Functions in PostgreSQL

Definition:
A window function in PostgreSQL performs a calculation across a set of
table rows that are somehow related to the current row. Unlike regular
aggregate functions, window functions do not collapse rows; they
return a value for every row and utilize an OVER() clause to define
the window of rows for calculations.

Syntax Example:
SELECT
    col1,
    col2,
    SUM(col3) OVER (
        PARTITION BY col1
        ORDER BY col2
        ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
    ) AS running_total
FROM
    my_table;

Common Use Cases:
1. Running totals or cumulative sums for each group (e.g., customer).
2. Calculating moving averages or running minimum/maximum values.
3. Ranking rows within partitions using ROW_NUMBER(), RANK(), or
   DENSE_RANK().
4. Calculating differences between current and previous or next row
   (using LAG() or LEAD()).
5. Percentile or N-tile analysis within groups.

Performance Implications and Pitfalls:
- Window functions may require sorting and/or scanning large data
  partitions, which can be resource intensive.
- Performance is heavily dependent on indexes that match the PARTITION
  BY and ORDER BY clauses used in the window.
- Using complex window frames or large partitions can slow down query
  execution and increase memory usage.
- Be careful with unbounded window frames over large data sets; they
  can result in high computational cost.
- Windowed aggregates do not filter or reduce output rows—use them when
  every row requires a calculated value based on its context.
*/

-- Calculate the running total of payments for each customer.
SELECT
    p.customer_id,
    p.payment_date,
    p.amount,
    sum(p.amount) OVER (
        PARTITION BY
            p.customer_id
        ORDER BY
            p.payment_date
        ROWS BETWEEN UNBOUNDED PRECEDING
        AND CURRENT ROW
    ) AS running_total
FROM
    payment AS p
ORDER BY
    p.customer_id,
    p.payment_date;

-- Calculate the time since the previous rental for each customer.
SELECT
    customer_id,
    rental_id,
    rental_date,
    rental_date - lag(rental_date) OVER (
        PARTITION BY customer_id ORDER BY rental_date
    ) AS since_prev_rental
FROM rental
ORDER BY customer_id, rental_date;

-- Calculate the percentage of total revenue for each film.
SELECT
    f.film_id,
    f.title,
    x.rev,
    round(100.0 * x.rev / sum(x.rev) OVER (), 2) AS pct_of_total
FROM
    (
        SELECT
            inventory.film_id,
            sum(p.amount) AS rev
        FROM
            payment AS p
        INNER JOIN rental ON p.rental_id = rental.rental_id
        INNER JOIN inventory ON rental.inventory_id = inventory.inventory_id
        GROUP BY
            inventory.film_id
    ) AS x
INNER JOIN film AS f ON x.film_id = f.film_id
ORDER BY
    x.rev DESC;
