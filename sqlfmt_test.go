package sqlfmt

import (
	"fmt"
	"testing"
)

func TestFormatSQL(t *testing.T) {

	type Test struct {
		SQL string
		WantSQL string
	}

	var testsData = []Test{
		{
			SQL: `SELECT count(*) count, winner, counter * (60 * 5) as counter FROM (SELECT winner, round(length / (60 * 5)) as counter FROM players WHERE build = $1 AND (hero = $2 OR region = $3)) GROUP BY winner, counter;`,
			WantSQL:`SELECT
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
		},
		{
			SQL:`INSERT INTO players(build, hero, region, winner, length) VALUES ($1, $2, $3, $4, $5);`,
			WantSQL: `INSERT
INTO
	players (build, hero, region, winner, length)
VALUES
	($1, $2, $3, $4, $5);`,
		},
		{
		SQL: `UPDATE players SET count = 0 WHERE build = $1 AND (hero = $2 OR region = $3) LIMIT 1;`,
		WantSQL: `UPDATE
	players
SET
	count = 0
WHERE
	build = $1 AND (hero = $2 OR region = $3)
LIMIT
	1;`,
		},
		{
		SQL:`SELECT * from test where a = 1`,
		WantSQL:`SELECT
	*
FROM
	test
WHERE
	a = 1;`,
		},
	}

	for i, item := range testsData {
		formatSQL, err := FormatSQL(item.SQL)
		if err != nil {
			t.Fatal(err)
		}
		if formatSQL != item.WantSQL {
			//os.WriteFile(fmt.Sprintf("%d.sql",i),[]byte(formatSQL),0644)
			t.Fatalf("format sql %d failed: %s",i,formatSQL)
		}
	}
}

func TestSimpleFormatSQL(t *testing.T) {
	formatSQL, err := FormatSQL("SELECT * FROM test WHERE A >= 4")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(formatSQL)
}


