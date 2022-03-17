// Copyright 2019 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package colexecutils

import (
	"context"
	"testing"

	"sqlfmt/cockroach/pkg/sql/colexecerror"
	"sqlfmt/cockroach/pkg/sql/colexecop"
	"sqlfmt/cockroach/pkg/sql/types"
	"sqlfmt/cockroach/pkg/util/cancelchecker"
	"sqlfmt/cockroach/pkg/util/leaktest"
	"sqlfmt/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/require"
)

// TestCancelChecker verifies that CancelChecker panics with appropriate error
// when the context is canceled.
func TestCancelChecker(t *testing.T) {
	defer leaktest.AfterTest(t)()
	defer log.Scope(t).Close(t)
	ctx, cancel := context.WithCancel(context.Background())
	typs := []*types.T{types.Int}
	batch := testAllocator.NewMemBatchWithMaxCapacity(typs)
	op := NewCancelChecker(colexecop.NewRepeatableBatchSource(testAllocator, batch, typs))
	op.Init(ctx)
	cancel()
	err := colexecerror.CatchVectorizedRuntimeError(func() {
		op.Next()
	})
	require.True(t, errors.Is(err, cancelchecker.QueryCanceledError))
}
