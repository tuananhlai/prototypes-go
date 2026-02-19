/*
GROUPING SETS in SQL

Definition:
GROUPING SETS is an extension of the GROUP BY clause that allows you to
specify multiple distinct groupings of data in a single query. Unlike
ROLLUP (which assumes a hierarchy) or CUBE (which generates all possible
combinations), GROUPING SETS lets you explicitly define exactly which
combinations of columns you want to aggregate.

Syntax Example:
SELECT
    col1,
    col2,
    SUM(col3)
FROM
    my_table
GROUP BY
    GROUPING SETS (
        (col1, col2),
        (col1),
        ()
    );

Use Cases:
1. Custom Aggregations: When you need specific multiple aggregates that
   don't fit a strict hierarchy (ROLLUP) or all-combinations (CUBE)
   pattern.
2. Cross-Tab Reports: Showing totals by different independent dimensions
   (e.g., total sales by Product and total sales by Region in the same
   result set).
3. Dashboard Data: Fetching multiple summary statistics in one database
   hit.

Why use GROUPING SETS:
It provides precise control over which aggregation levels are computed,
avoiding the generation of unnecessary subtotals that ROLLUP or CUBE
might produce.

Performance Implications and Pitfalls:
- Efficiency: Like ROLLUP, it is typically more efficient than multiple
  UNION ALL queries. The database engine can optimize the aggregation
  plan to share sort or hash operations.
- Complexity: The result set can be complex to consume because different
  rows represent different grouping granularities.
- GROUPING() function: Use the GROUPING() or GROUPING_ID() functions to
  identify which grouping set each row belongs to.
*/

SELECT
    drivers.surname AS driver,
    constructors.name AS constructor,
    sum(results.points) AS points
FROM
    results
INNER JOIN races ON results.race_id = races.race_id
INNER JOIN drivers ON results.driver_id = drivers.driver_id
INNER JOIN constructors ON results.constructor_id = constructors.constructor_id
WHERE
    races.date >= '1978-01-01'
    AND races.date < date '1978-01-01' + interval '1 year'
GROUP BY
    GROUPING SETS ((drivers.surname), (constructors.name))
HAVING
    sum(points) > 20
ORDER BY
    constructors.name IS NOT NULL,
    drivers.surname IS NOT NULL,
    points DESC;
