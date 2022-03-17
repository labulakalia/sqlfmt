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

	"sqlfmt/cockroach/pkg/testutils/buildutil"
)

func TestNoLinkForbidden(t *testing.T) {
	// Prohibit introducing any new dependencies into this package since it
	// should be very lightweight.
	buildutil.VerifyNoImports(t,
		"sqlfmt/cockroach/pkg/sql/colexec/colexecargs", true,
		nil /* forbiddenPkgs */, nil, /* forbiddenPrefixes */
		// allowlist:
		"sqlfmt/cockroach/pkg/col/coldata",
		"sqlfmt/cockroach/pkg/sql/colcontainer",
		"sqlfmt/cockroach/pkg/sql/colexecop",
		"sqlfmt/cockroach/pkg/sql/execinfra",
		"sqlfmt/cockroach/pkg/sql/execinfrapb",
		"sqlfmt/cockroach/pkg/sql/types",
		"sqlfmt/cockroach/pkg/util/mon",
		"github.com/marusama/semaphore",
	)
}
