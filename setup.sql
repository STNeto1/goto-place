DROP TABLE IF EXISTS points;

CREATE TABLE points (
    x int not null,
    y int not null,
    color text not null,

    PRIMARY KEY (x, y)
);


WITH RECURSIVE x_series AS (
  SELECT 0 AS x
  UNION ALL
  SELECT x + 1
  FROM x_series
  WHERE x < 9
),
y_series AS (
  SELECT 0 AS y
  UNION ALL
  SELECT y + 1
  FROM y_series
  WHERE y < 9
)
INSERT INTO points (x, y, color)
SELECT x, y, colors.color
FROM x_series
CROSS JOIN y_series
CROSS JOIN (SELECT '#333' AS color) AS colors
ORDER BY x, y;
