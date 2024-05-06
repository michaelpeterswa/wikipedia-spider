SELECT
    *
FROM
    wikipediascraper.links
WHERE
    link = $1
LIMIT
    1;