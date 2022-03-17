// Copyright 2021 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package colexecspan

import (
	"flag"
	"fmt"
	"os"
	"testing"

	"sqlfmt/cockroach/pkg/col/coldata"
	"sqlfmt/cockroach/pkg/sql/colexec/colexectestutils"
	"sqlfmt/cockroach/pkg/sql/colexecerror"
	"sqlfmt/cockroach/pkg/testutils/skip"
	"sqlfmt/cockroach/pkg/util/randutil"
)

func TestMain(m *testing.M) {
	randutil.SeedForTests()
	os.Exit(func() int {
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
