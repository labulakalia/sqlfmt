// Copyright 2021 The Cockroach Authors.
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
	"sqlfmt/cockroach/pkg/sql/colexec/colexectestutils"
	"sqlfmt/cockroach/pkg/sql/execinfrapb"
	"sqlfmt/cockroach/pkg/sql/types"
)

type sortTestCase struct {
	description string
	tuples      colexectestutils.Tuples
	expected    colexectestutils.Tuples
	typs        []*types.T
	ordCols     []execinfrapb.Ordering_Column
	matchLen    int
	k           uint64
}
