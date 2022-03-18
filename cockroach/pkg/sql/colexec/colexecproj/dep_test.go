// Copyright 2021 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package colexecproj

import (
	"testing"

	"github.com/labulakalia/sqlfmt/cockroach/pkg/testutils/buildutil"
)

func TestNoLinkForbidden(t *testing.T) {
	buildutil.VerifyNoImports(t,
		"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/colexec/colexecproj", true,
		[]string{
			"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/colexec",
			"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/colexec/colexecagg",
			"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/colexec/colexechash",
			"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/colexec/colexecjoin",
			"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/colexec/colexecbase",
			"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/colexec/colexecsel",
			"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/colexec/colexecwindow",
		}, nil,
	)
}
