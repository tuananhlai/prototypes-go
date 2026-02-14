SELECT
	drivers.surname AS driver,
	constructors.name AS constructor,
	sum(points) AS points
FROM
	results
	JOIN races USING (race_id)
	JOIN drivers USING (driver_id)
	JOIN constructors USING (constructor_id)
WHERE
	date >= '1978-01-01'
	AND date < date '1978-01-01' + interval '1 year'
GROUP BY
	GROUPING sets ((drivers.surname), (constructors.name))
HAVING
	sum(points) > 20
ORDER BY
	constructors.name IS NOT NULL,
	drivers.surname IS NOT NULL,
	points DESC;
