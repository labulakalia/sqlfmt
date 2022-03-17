// Copyright 2020 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package sql_test

import (
	"context"
	"testing"

	"sqlfmt/cockroach/pkg/base"
	"sqlfmt/cockroach/pkg/keys"
	"sqlfmt/cockroach/pkg/roachpb"
	"sqlfmt/cockroach/pkg/testutils/serverutils"
	"sqlfmt/cockroach/pkg/testutils/sqlutils"
	"sqlfmt/cockroach/pkg/util/leaktest"
	"sqlfmt/cockroach/pkg/util/log"
	"github.com/stretchr/testify/require"
)

// TestDestroyTenantSynchronous tests that destroying a tenant synchronously
// with crdb_internal.destroy_tenant(<ten_id>, true) works.
func TestDestroyTenantSynchronous(t *testing.T) {
	defer leaktest.AfterTest(t)()
	defer log.Scope(t).Close(t)

	ctx := context.Background()
	s, sqlDB, kvDB := serverutils.StartServer(t, base.TestServerArgs{})
	defer s.Stopper().Stop(ctx)

	tenantID := roachpb.MakeTenantID(10)
	codec := keys.MakeSQLCodec(tenantID)
	const tenantStateQuery = `
SELECT id, active FROM system.tenants WHERE id = 10
`
	checkKVsExistForTenant := func(t *testing.T, shouldExist bool) {
		rows, err := kvDB.Scan(
			ctx, codec.TenantPrefix(), codec.TenantPrefix().PrefixEnd(),
			1, /* maxRows */
		)
		require.NoError(t, err)
		require.Equal(t, shouldExist, len(rows) > 0)
	}

	tdb := sqlutils.MakeSQLRunner(sqlDB)

	// Create the tenant, make sure it has data and state.
	tdb.Exec(t, "SELECT crdb_internal.create_tenant(10)")
	tdb.CheckQueryResults(t, tenantStateQuery, [][]string{{"10", "true"}})
	checkKVsExistForTenant(t, true /* shouldExist */)

	// Destroy the tenant, make sure it does not have data and state.
	tdb.Exec(t, "SELECT crdb_internal.destroy_tenant(10, true)")
	tdb.CheckQueryResults(t, tenantStateQuery, [][]string{})
	checkKVsExistForTenant(t, false /* shouldExist */)
}
