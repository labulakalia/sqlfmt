// Copyright 2021 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package colexecargs

import (
	"testing"

	"github.com/labulakalia/sqlfmt/cockroach/pkg/testutils/buildutil"
)

func TestNoLinkForbidden(t *testing.T) {
	// Prohibit introducing any new dependencies into this package since it
	// should be very lightweight.
	buildutil.VerifyNoImports(t,
		"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/colexec/colexecargs", true,
		nil /* forbiddenPkgs */, nil, /* forbiddenPrefixes */
		// allowlist:
		"github.com/labulakalia/sqlfmt/cockroach/pkg/col/coldata",
		"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/colcontainer",
		"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/colexecop",
		"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/execinfra",
		"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/execinfrapb",
		"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/types",
		"github.com/labulakalia/sqlfmt/cockroach/pkg/util/mon",
		"github.com/marusama/semaphore",
	)
}
