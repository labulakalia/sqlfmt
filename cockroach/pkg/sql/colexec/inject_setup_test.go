// Copyright 2020 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package colexec_test

import (
	"sqlfmt/cockroach/pkg/sql/colexec/colbuilder"
	"sqlfmt/cockroach/pkg/sql/colexec/colexecargs"
	"sqlfmt/cockroach/pkg/sql/colexec/colexecbase"
	"sqlfmt/cockroach/pkg/sql/colexec/colexectestutils"
)

func init() {
	// Inject a testing helper for NewColOperator so colexec tests can
	// use NewColOperator without an import cycle.
	colexecargs.TestNewColOperator = colbuilder.NewColOperator
	// Inject a testing helper for OrderedDistinctColsToOperators so colexec tests
	// can use OrderedDistinctColsToOperators without an import cycle.
	colexectestutils.OrderedDistinctColsToOperators = colexecbase.OrderedDistinctColsToOperators
}
