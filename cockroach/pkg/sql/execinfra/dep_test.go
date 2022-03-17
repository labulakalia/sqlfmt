// Copyright 2019 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package execinfra

import (
	"testing"

	"sqlfmt/cockroach/pkg/testutils/buildutil"
	"sqlfmt/cockroach/pkg/util/leaktest"
)

func TestNoLinkForbidden(t *testing.T) {
	defer leaktest.AfterTest(t)()

	buildutil.VerifyNoImports(t,
		"sqlfmt/cockroach/pkg/sql/execinfra", true,
		[]string{
			"sqlfmt/cockroach/pkg/sql/colexec",
			"sqlfmt/cockroach/pkg/sql/colflow",
			"sqlfmt/cockroach/pkg/sql/flowinfra",
			"sqlfmt/cockroach/pkg/sql/rowexec",
			"sqlfmt/cockroach/pkg/sql/rowflow",
		}, nil,
	)
}
