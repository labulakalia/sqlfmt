// Copyright 2019 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package colexec

import (
	"context"
	"testing"

	"sqlfmt/cockroach/pkg/col/coldata"
	"sqlfmt/cockroach/pkg/col/coldatatestutils"
	"sqlfmt/cockroach/pkg/sql/colexec/colexecargs"
	"sqlfmt/cockroach/pkg/sql/colexec/colexectestutils"
	"sqlfmt/cockroach/pkg/sql/colexecop"
	"sqlfmt/cockroach/pkg/sql/execinfrapb"
	"sqlfmt/cockroach/pkg/sql/types"
	"sqlfmt/cockroach/pkg/util/leaktest"
	"sqlfmt/cockroach/pkg/util/log"
	"sqlfmt/cockroach/pkg/util/randutil"
	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/require"
)

func TestSerialUnorderedSynchronizer(t *testing.T) {
	defer leaktest.AfterTest(t)()
	defer log.Scope(t).Close(t)

	ctx := context.Background()
	rng, _ := randutil.NewTestRand()
	const numInputs = 3
	const numBatches = 4

	typs := []*types.T{types.Int}
	inputs := make([]colexecargs.OpWithMetaInfo, numInputs)
	for i := range inputs {
		batch := coldatatestutils.RandomBatch(testAllocator, rng, typs, coldata.BatchSize(), 0 /* length */, rng.Float64())
		source := colexecop.NewRepeatableBatchSource(testAllocator, batch, typs)
		source.ResetBatchesToReturn(numBatches)
		inputIdx := i
		inputs[i].Root = source
		inputs[i].MetadataSources = []colexecop.MetadataSource{
			colexectestutils.CallbackMetadataSource{
				DrainMetaCb: func() []execinfrapb.ProducerMetadata {
					return []execinfrapb.ProducerMetadata{{Err: errors.Errorf("input %d test-induced metadata", inputIdx)}}
				},
			},
		}
	}
	s := NewSerialUnorderedSynchronizer(inputs)
	s.Init(ctx)
	resultBatches := 0
	for {
		b := s.Next()
		if b.Length() == 0 {
			require.Equal(t, len(inputs), len(s.DrainMeta()))
			break
		}
		resultBatches++
	}
	require.Equal(t, numInputs*numBatches, resultBatches)
}
