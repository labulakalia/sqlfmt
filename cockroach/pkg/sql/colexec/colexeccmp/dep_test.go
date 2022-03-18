// Copyright 2021 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package colexeccmp

import (
	"testing"

	"github.com/labulakalia/sqlfmt/cockroach/pkg/testutils/buildutil"
)

func TestNoLinkForbidden(t *testing.T) {
	// Prohibit introducing any new dependencies into this package since it
	// should be very lightweight.
	buildutil.VerifyNoImports(t,
		"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/colexec/colexeccmp", true,
		nil /* forbiddenPkgs */, nil, /* forbiddenPrefixes */
		// allowlist:
		"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/sem/tree",
		"github.com/cockroachdb/errors",
	)
}
