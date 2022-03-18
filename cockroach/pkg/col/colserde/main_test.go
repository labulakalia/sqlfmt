// Copyright 2019 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package colserde_test

import (
	"context"
	"os"
	"testing"

	"github.com/labulakalia/sqlfmt/cockroach/pkg/col/coldata"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/col/coldataext"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/settings/cluster"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/colmem"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/execinfra"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/util/mon"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/util/randutil"
)

var (
	// testAllocator is a colexec.Allocator with an unlimited budget for use in
	// tests.
	testAllocator     *colmem.Allocator
	testColumnFactory coldata.ColumnFactory

	// testMemMonitor and testMemAcc are a test monitor with an unlimited budget
	// and a memory account bound to it for use in tests.
	testMemMonitor *mon.BytesMonitor
	testMemAcc     *mon.BoundAccount
)

func TestMain(m *testing.M) {
	randutil.SeedForTests()
	os.Exit(func() int {
		ctx := context.Background()
		st := cluster.MakeTestingClusterSettings()
		testMemMonitor = execinfra.NewTestMemMonitor(ctx, st)
		defer testMemMonitor.Stop(ctx)
		memAcc := testMemMonitor.MakeBoundAccount()
		testMemAcc = &memAcc
		testColumnFactory = coldataext.NewExtendedColumnFactory(nil /* evalCtx */)
		testAllocator = colmem.NewAllocator(ctx, testMemAcc, testColumnFactory)
		defer testMemAcc.Close(ctx)
		return m.Run()
	}())
}
