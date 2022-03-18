// Copyright 2017 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package physicalplan_test

import (
	"context"
	"testing"

	"github.com/labulakalia/sqlfmt/cockroach/pkg/base"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/keys"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/kv"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/kv/kvclient/kvcoord"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/roachpb"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/catalog/desctestutils"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/sem/tree"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/testutils/physicalplanutils"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/testutils/serverutils"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/testutils/sqlutils"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/util/leaktest"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/util/log"
)

func TestFakeSpanResolver(t *testing.T) {
	defer leaktest.AfterTest(t)()
	defer log.Scope(t).Close(t)
	ctx := context.Background()

	tc := serverutils.StartNewTestCluster(t, 3, base.TestClusterArgs{})
	defer tc.Stopper().Stop(ctx)

	sqlutils.CreateTable(
		t, tc.ServerConn(0), "t",
		"k INT PRIMARY KEY, v INT",
		100,
		func(row int) []tree.Datum {
			return []tree.Datum{
				tree.NewDInt(tree.DInt(row)),
				tree.NewDInt(tree.DInt(row * row)),
			}
		},
	)

	resolver := physicalplanutils.FakeResolverForTestCluster(tc)

	db := tc.Server(0).DB()

	txn := kv.NewTxn(ctx, db, tc.Server(0).NodeID())
	it := resolver.NewSpanResolverIterator(txn)

	tableDesc := desctestutils.TestingGetPublicTableDescriptor(db, keys.SystemSQLCodec, "test", "t")
	primIdxValDirs := catalogkeys.IndexKeyValDirs(tableDesc.GetPrimaryIndex())

	span := tableDesc.PrimaryIndexSpan(keys.SystemSQLCodec)

	// Make sure we see all the nodes. It will not always happen (due to
	// randomness) but it should happen most of the time.
	for attempt := 0; attempt < 10; attempt++ {
		nodesSeen := make(map[roachpb.NodeID]struct{})
		it.Seek(ctx, span, kvcoord.Ascending)
		lastKey := span.Key
		for {
			if !it.Valid() {
				t.Fatal(it.Error())
			}
			desc := it.Desc()
			rinfo, err := it.ReplicaInfo(ctx)
			if err != nil {
				t.Fatal(err)
			}

			prettyStart := keys.PrettyPrint(primIdxValDirs, desc.StartKey.AsRawKey())
			prettyEnd := keys.PrettyPrint(primIdxValDirs, desc.EndKey.AsRawKey())
			t.Logf("%d %s %s", rinfo.NodeID, prettyStart, prettyEnd)

			if !lastKey.Equal(desc.StartKey.AsRawKey()) {
				t.Errorf("unexpected start key %s, should be %s", prettyStart, keys.PrettyPrint(primIdxValDirs, span.Key))
			}
			if !desc.StartKey.Less(desc.EndKey) {
				t.Errorf("invalid range %s to %s", prettyStart, prettyEnd)
			}
			lastKey = desc.EndKey.AsRawKey()
			nodesSeen[rinfo.NodeID] = struct{}{}

			if !it.NeedAnother() {
				break
			}
			it.Next(ctx)
		}

		if !lastKey.Equal(span.EndKey) {
			t.Errorf("last key %s, should be %s", keys.PrettyPrint(primIdxValDirs, lastKey), keys.PrettyPrint(primIdxValDirs, span.EndKey))
		}

		if len(nodesSeen) == tc.NumServers() {
			// Saw all the nodes.
			break
		}
		t.Logf("not all nodes seen; retrying")
	}
}
