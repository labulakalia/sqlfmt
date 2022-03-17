// Copyright 2021 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package colexecsel

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
	"sqlfmt/cockroach/pkg/sql/sem/tree"
	"sqlfmt/cockroach/pkg/testutils/skip"
	"sqlfmt/cockroach/pkg/util/randutil"
)

// testAllocator is an Allocator with an unlimited budget for use in tests.
var testAllocator *colmem.Allocator

func TestMain(m *testing.M) {
	randutil.SeedForTests()
	os.Exit(func() int {
		ctx := context.Background()
		st := cluster.MakeTestingClusterSettings()
		testMemMonitor := execinfra.NewTestMemMonitor(ctx, st)
		defer testMemMonitor.Stop(ctx)
		memAcc := testMemMonitor.MakeBoundAccount()
		evalCtx := tree.MakeTestingEvalContext(st)
		testColumnFactory := coldataext.NewExtendedColumnFactory(&evalCtx)
		testAllocator = colmem.NewAllocator(ctx, &memAcc, testColumnFactory)
		defer memAcc.Close(ctx)

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
