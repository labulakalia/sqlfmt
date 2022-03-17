package sqlfmt

import (
	"testing"
)

func TestFormatSQL(t *testing.T) {
	sqls := []string{
		`SELECT count(*) count, winner, counter * (60 * 5) as counter FROM (SELECT winner, round(length / (60 * 5)) as counter FROM players WHERE build = $1 AND (hero = $2 OR region = $3)) GROUP BY winner, counter;`,
		`INSERT INTO players(build, hero, region, winner, length) VALUES ($1, $2, $3, $4, $5);`,
		`UPDATE players SET count = 0 WHERE build = $1 AND (hero = $2 OR region = $3) LIMIT 1;`,
	}

	res := []string{
`SELECT
	count(*) AS count, winner, counter * 60 * 5 AS counter
FROM
	(
		SELECT
			winner, round(length / (60 * 5)) AS counter
		FROM
			players
		WHERE
			build = $1 AND (hero = $2 OR region = $3)
	)
GROUP BY
	winner, counter;`,
`INSERT
INTO
	players (build, hero, region, winner, length)
VALUES
	($1, $2, $3, $4, $5);`,
`UPDATE
	players
SET
	count = 0
WHERE
	build = $1 AND (hero = $2 OR region = $3)
LIMIT
	1;`,
	}
	for i, sql := range sqls {
		formatSQL, err := FormatSQL(sql)
		if err != nil {
			t.Fatal(err)
		}
		if formatSQL != res[i] {
			t.Fatalf("format sql %d failed: %s",i,formatSQL)
		}
	}

}
