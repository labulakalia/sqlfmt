// Copyright 2021 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package scdecomp_test

import (
	"context"
	gosql "database/sql"
	"testing"

	"github.com/labulakalia/sqlfmt/cockroach/pkg/base"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/schemachanger/scrun"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/schemachanger/sctest"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/testutils"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/testutils/testcluster"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/util/leaktest"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/util/log"
)

func TestDecomposeToElements(t *testing.T) {
	defer leaktest.AfterTest(t)()
	defer log.Scope(t).Close(t)
	ctx := context.Background()

	newCluster := func(t *testing.T, knobs *scrun.TestingKnobs) (_ *gosql.DB, cleanup func()) {
		tc := testcluster.StartTestCluster(t, 1, base.TestClusterArgs{})
		return tc.ServerConn(0), func() { tc.Stopper().Stop(ctx) }
	}

	sctest.DecomposeToElements(t, testutils.TestDataPath(t), newCluster)
}
