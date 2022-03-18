// Copyright 2020 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package colexecop

import (
	"testing"

	"github.com/labulakalia/sqlfmt/cockroach/pkg/testutils/buildutil"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/util/leaktest"
)

func TestNoLinkForbidden(t *testing.T) {
	defer leaktest.AfterTest(t)()

	buildutil.VerifyNoImports(t,
		"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/colexecop", true,
		[]string{
			"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/colcontainer",
			"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/colexec",
			"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/colflow",
			"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/rowexec",
			"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/rowflow",
		}, nil,
	)
}
