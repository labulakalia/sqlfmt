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

	"sqlfmt/cockroach/pkg/keys"
	"sqlfmt/cockroach/pkg/kv"
	"sqlfmt/cockroach/pkg/sql/catalog/catalogkeys"
	"sqlfmt/cockroach/pkg/sql/tests"
	"sqlfmt/cockroach/pkg/testutils/serverutils"
	"sqlfmt/cockroach/pkg/util/leaktest"
	"sqlfmt/cockroach/pkg/util/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLookupNamespaceID tests the lookup namespace id.
func TestLookupNamespaceID(t *testing.T) {
	defer leaktest.AfterTest(t)()
	defer log.Scope(t).Close(t)

	ctx := context.Background()
	params, _ := tests.CreateTestServerParams()
	s, sqlDB, kvDB := serverutils.StartServer(t, params)
	defer s.Stopper().Stop(ctx)

	err := kvDB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		return txn.Put(
			ctx,
			catalogkeys.MakeObjectNameKey(keys.SystemSQLCodec, 999, 1000, "bob"),
			9999,
		)
	})
	require.NoError(t, err)

	// Assert the row exists in the database.
	var id int64
	err = sqlDB.QueryRow(
		`SELECT crdb_internal.get_namespace_id($1, $2, $3)`,
		999,
		1000,
		"bob",
	).Scan(&id)
	require.NoError(t, err)
	assert.Equal(t, int64(9999), id)
}
