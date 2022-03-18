// Copyright 2021 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package migrations

import (
	"context"

	"github.com/labulakalia/sqlfmt/cockroach/pkg/clusterversion"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/jobs"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/keys"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/migration"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/catalog/systemschema"
)

const addAvgSizeCol = `
ALTER TABLE system.table_statistics 
ADD COLUMN IF NOT EXISTS "avgSize" INT8 NOT NULL DEFAULT (INT8 '0')
`

func alterSystemTableStatisticsAddAvgSize(
	ctx context.Context, cs clusterversion.ClusterVersion, d migration.TenantDeps, _ *jobs.Job,
) error {
	op := operation{
		name:           "add-table-statistics-avgSize-col",
		schemaList:     []string{"total_consumption"},
		query:          addAvgSizeCol,
		schemaExistsFn: hasColumn,
	}
	if err := migrateTable(ctx, cs, d, op, keys.TableStatisticsTableID, systemschema.TableStatisticsTable); err != nil {
		return err
	}
	return nil
}
