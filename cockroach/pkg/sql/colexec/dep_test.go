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
	"testing"

	"sqlfmt/cockroach/pkg/testutils/buildutil"
	"sqlfmt/cockroach/pkg/util/leaktest"
	"sqlfmt/cockroach/pkg/util/log"
)

func TestNoLinkForbidden(t *testing.T) {
	defer leaktest.AfterTest(t)()
	defer log.Scope(t).Close(t)

	buildutil.VerifyNoImports(t,
		"sqlfmt/cockroach/pkg/sql/colexec", true,
		[]string{
			"sqlfmt/cockroach/pkg/sql/colflow",
			"sqlfmt/cockroach/pkg/sql/rowexec",
			"sqlfmt/cockroach/pkg/sql/rowflow",
		}, nil,
	)
}
