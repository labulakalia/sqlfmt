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
	"testing"

	"sqlfmt/cockroach/pkg/testutils/buildutil"
)

func TestNoLinkForbidden(t *testing.T) {
	buildutil.VerifyNoImports(t,
		"sqlfmt/cockroach/pkg/sql/colexec/colexecbase", true,
		[]string{
			"sqlfmt/cockroach/pkg/sql/colexec",
			"sqlfmt/cockroach/pkg/sql/colexec/colexecagg",
			"sqlfmt/cockroach/pkg/sql/colexec/colexechash",
			"sqlfmt/cockroach/pkg/sql/colexec/colexecproj",
			"sqlfmt/cockroach/pkg/sql/colexec/colexecsel",
		}, nil,
	)
}
