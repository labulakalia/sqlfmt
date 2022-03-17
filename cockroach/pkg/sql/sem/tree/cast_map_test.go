// Copyright 2021 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package tree_test

import (
	"testing"

	"sqlfmt/cockroach/pkg/settings/cluster"
	"sqlfmt/cockroach/pkg/sql/faketreeeval"
	"sqlfmt/cockroach/pkg/sql/pgwire/pgcode"
	"sqlfmt/cockroach/pkg/sql/pgwire/pgerror"
	"sqlfmt/cockroach/pkg/sql/randgen"
	"sqlfmt/cockroach/pkg/sql/sem/tree"
	"sqlfmt/cockroach/pkg/sql/types"
	"sqlfmt/cockroach/pkg/util/leaktest"
	"sqlfmt/cockroach/pkg/util/log"
	"sqlfmt/cockroach/pkg/util/randutil"
	"github.com/lib/pq/oid"
)

// TestCastMap tests that every cast in tree.castMap can be performed by
// PerformCast.
func TestCastMap(t *testing.T) {
	defer leaktest.AfterTest(t)()
	defer log.Scope(t).Close(t)

	evalCtx := tree.MakeTestingEvalContext(cluster.MakeTestingClusterSettings())
	rng, _ := randutil.NewTestRand()
	evalCtx.Planner = &faketreeeval.DummyEvalPlanner{}

	tree.ForEachCast(func(src, tgt oid.Oid) {
		srcType := types.OidToType[src]
		tgtType := types.OidToType[tgt]
		srcDatum := randgen.RandDatum(rng, srcType, false /* nullOk */)

		// TODO(mgartner): We do not allow casting a negative integer to bit
		// types with unbounded widths. Until we add support for this, we
		// ensure that the srcDatum is positive.
		if srcType.Family() == types.IntFamily && tgtType.Family() == types.BitFamily {
			srcVal := *srcDatum.(*tree.DInt)
			if srcVal < 0 {
				srcDatum = tree.NewDInt(-srcVal)
			}
		}

		_, err := tree.PerformCast(&evalCtx, srcDatum, tgtType)
		// If the error is a CannotCoerce error, then PerformCast does not
		// support casting from src to tgt. The one exception is negative
		// integers to bit types which return the same error code (see the TODO
		// above).
		if err != nil && pgerror.HasCandidateCode(err) && pgerror.GetPGCode(err) == pgcode.CannotCoerce {
			t.Errorf("cast from %s to %s failed: %s", srcType, tgtType, err)
		}
	})
}
