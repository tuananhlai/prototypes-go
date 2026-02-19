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
    drivers.surname IN ('Prost', 'Senna')
GROUP BY
    cube (drivers.surname, constructors.name)
