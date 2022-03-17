// Copyright 2021 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package colexecbase_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	"sqlfmt/cockroach/pkg/col/coldata"
	"sqlfmt/cockroach/pkg/col/coldataext"
	"sqlfmt/cockroach/pkg/settings/cluster"
	"sqlfmt/cockroach/pkg/sql/colexec/colexectestutils"
	"sqlfmt/cockroach/pkg/sql/colexecerror"
	"sqlfmt/cockroach/pkg/sql/colmem"
	"sqlfmt/cockroach/pkg/sql/execinfra"
	"sqlfmt/cockroach/pkg/sql/faketreeeval"
	"sqlfmt/cockroach/pkg/sql/sem/tree"
	"sqlfmt/cockroach/pkg/testutils/skip"
	"sqlfmt/cockroach/pkg/util/mon"
	"sqlfmt/cockroach/pkg/util/randutil"
)

var (
	// testAllocator is an Allocator with an unlimited budget for use in tests.
	testAllocator     *colmem.Allocator
	testColumnFactory coldata.ColumnFactory

	// testMemMonitor and testMemAcc are a test monitor with an unlimited budget
	// and a memory account bound to it for use in tests.
	testMemMonitor *mon.BytesMonitor
	testMemAcc     *mon.BoundAccount

	// testDiskMonitor and testDiskAcc are a test monitor with an unlimited budget
	// and a disk account bound to it for use in tests.
	testDiskMonitor *mon.BytesMonitor
	testDiskAcc     *mon.BoundAccount
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
		evalCtx := tree.MakeTestingEvalContext(st)
		evalCtx.Planner = &faketreeeval.DummyEvalPlanner{}
		testColumnFactory = coldataext.NewExtendedColumnFactory(&evalCtx)
		testAllocator = colmem.NewAllocator(ctx, testMemAcc, testColumnFactory)
		defer testMemAcc.Close(ctx)

		testDiskMonitor = execinfra.NewTestDiskMonitor(ctx, st)
		defer testDiskMonitor.Stop(ctx)
		diskAcc := testDiskMonitor.MakeBoundAccount()
		testDiskAcc = &diskAcc
		defer testDiskAcc.Close(ctx)

		flag.Parse()
		if !skip.UnderBench() {
			// (If we're running benchmarks, don't set a random batch size.)
			randomBatchSize := colexectestutils.GenerateBatchSize()
			fmt.Printf("coldata.BatchSize() is set to %d\n", randomBatchSize)
			if err := coldata.SetBatchSizeForTests(randomBatchSize); err != nil {
				colexecerror.InternalError(err)
			}
		}
		return m.Run()
	}())
}
