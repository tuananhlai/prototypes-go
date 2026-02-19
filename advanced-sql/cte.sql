WITH
accidents AS (
    SELECT
        extract(
            YEAR
            FROM
            races.date
        ) AS season,
        count(*) AS participants,
        count(*) FILTER (
            WHERE
            status.status = 'Accident'
        ) AS accidents
    FROM
        results
    INNER JOIN status ON results.status_id = status.status_id
    INNER JOIN races ON results.race_id = races.race_id
    GROUP BY
        season
)

SELECT
    season,
    round(100.0 * accidents / participants, 2) AS pct,
    repeat(text '■', ceil(100 * accidents / participants)::int) AS bar
FROM
    accidents
WHERE
    season BETWEEN 1974 AND 1990
ORDER BY
    season;
