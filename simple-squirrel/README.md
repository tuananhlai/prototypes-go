# Pagila SQL Training: 50 Exercises

Below is the list of exercises designed to cover the full spectrum of SQL and PostgreSQL features using the `pagila` database.

---

### Level 1: Very Easy (Basic Selection & Filtering)
1.  **Select All**: Retrieve all columns from the `actor` table.
2.  **Projection**: List only the `first_name` and `last_name` of every actor.
3.  **Simple Filter**: Find all films with a `rental_rate` of 0.99.
4.  **Equality Check**: List movie titles that have a `length` of exactly 90 minutes.
5.  **Distinct Values**: Get a list of unique film ratings (e.g., G, PG, R) from the `film` table.
6.  **Boolean Logic**: Find the email addresses of all active customers (`active = 1`).
7.  **Specific Year**: List all films released in the year 2006.
8.  **Ordering & Limits**: Show the first 10 records from the `payment` table, ordered by `payment_date`.
9.  **String Matching**: Find all actors whose first name is 'NICK'.
10. **Basic Count**: Get the total count of films available in the `inventory` table.

---

### Level 2: Easy (Joins & Aggregations)
11. **Inner Join**: Join `customer` and `address` to show each customer's name and phone number.
12. **Summation**: Calculate the total revenue generated from all payments in the database.
13. **Average**: Find the average length of all films.
14. **Grouped Count**: Show the count of films available for each rating category.
15. **Range Filtering**: List all films that have a `length` between 60 and 120 minutes.
16. **Max Value**: Find the maximum `replacement_cost` among all films.
17. **Multi-Table Join**: Join `film`, `film_category`, and `category` to display film titles and their category names.
18. **Filtered Aggregation**: Count how many rentals were processed by the staff member with `staff_id = 1`.
19. **Wildcard Search**: Show all films with titles starting with the letter 'W'.
20. **Having Clause**: Use `GROUP BY` and `HAVING` to find categories with more than 60 films.

---

### Level 3: Intermediate (Subqueries & String/Date Logic)
21. **Concatenation**: Concatenate `first_name` and `last_name` into a single column `full_name` for all actors.
22. **Conditional Logic**: Use a `CASE` statement to label films as 'Short' (<60m), 'Medium' (60-120m), or 'Long' (>120m).
23. **Inclusion/Exclusion**: Find all films that do *not* have a rating of 'R' or 'NC-17'.
24. **Date Extraction**: Extract the month and year from the `rental_date` for all rentals.
25. **Scalar Subquery**: Find actors who have the same first name as 'ED' using a subquery.
26. **Deep Joins**: List customers who live in 'Canada' (requires joining `customer`, `address`, `city`, and `country`).
27. **Set Operations**: Use `UNION` to combine the first names of all actors and all customers.
28. **Correlated Subquery**: Find the 10 most recent rentals and include the titles of the films rented.
29. **Null Handling**: Use `COALESCE` to display 'Not Returned' for any rentals where `return_date` is null.
30. **Top N per Group**: List the top 5 customers by total payment amount.

---

### Level 4: Hard (CTEs & Window Functions)
31. **Common Table Expressions**: Use a CTE to find the total revenue generated per film category.
32. **Ranking**: Use the `RANK()` window function to rank films by length within each rating category.
33. **Running Totals**: Calculate the "running total" of payments for each customer, ordered by payment date.
34. **Self-Joins**: Identify customers who have rented the same film more than once.
35. **Full-Text Search**: Find films where the description contains both 'Drama' and 'Student' (use the `fulltext` column or `tsvector` features).
36. **Aggregate Windowing**: Identify the date of the first and last rental for every customer.
37. **Time Series**: Calculate the monthly revenue for the year 2007.
38. **Left Joins & Nulls**: List all films that are currently "out" (rented but not yet returned) with customer names.
39. **Row Numbering**: Use `ROW_NUMBER()` to select only the most recent rental for every customer.
40. **Optimization**: Use `EXPLAIN ANALYZE` on a complex join between 5 tables to review the query execution plan.

---

### Level 5: Very Hard (DML, DDL & Advanced Postgres Features)
41. **Statistical Functions**: Use `PERCENT_RANK()` to find films in the top 5% of the longest movies.
42. **Views**: Create a View named `revenue_by_store` that sums payments per store.
43. **Bulk DML**: Write an `INSERT INTO ... SELECT` statement to move rentals older than a specific date into a `rental_archive` table.
44. **Lateral Joins**: Use `CROSS JOIN LATERAL` to find the 2 most expensive films for every category.
45. **Growth Analytics**: Calculate the month-over-month growth percentage for total revenue using `LAG()`.
46. **Schema Inspection**: Write a query against `information_schema.columns` to find all tables that contain a `last_update` column.
47. **Check Constraints**: Use `ALTER TABLE` to add a constraint to the `rental` table ensuring `return_date` is after `rental_date`.
48. **Transaction Management**: Demonstrate a transaction (`BEGIN`, `COMMIT`) that adds a new customer and their initial address record.
49. **Trigger Functions**: Create a trigger function to update the `last_update` timestamp automatically on any change to the `film` table.
50. **Recursive CTEs**: (If applicable to the schema) Find a chain of actors who have worked together, or simulate a hierarchical category search.
