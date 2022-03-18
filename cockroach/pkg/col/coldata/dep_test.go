// Copyright 2019 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package coldata

import (
	"testing"

	"github.com/labulakalia/sqlfmt/cockroach/pkg/testutils/buildutil"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/util/leaktest"
)

func TestNoLinkForbidden(t *testing.T) {
	defer leaktest.AfterTest(t)()

	buildutil.VerifyNoImports(t,
		"github.com/labulakalia/sqlfmt/cockroach/pkg/col/coldata", true,
		[]string{
			"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/catalog",
			"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/sem/tree",
		}, nil,
	)
}
