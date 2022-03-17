// Copyright 2022 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package main

import (
	"testing"

	"sqlfmt/cockroach/pkg/testutils/buildutil"
)

func TestNoLinkForbidden(t *testing.T) {
	buildutil.VerifyNoImports(t,
		"sqlfmt/cockroach/pkg/sql/colexec/execgen/cmd/execgen", true,
		[]string{
			"sqlfmt/cockroach/pkg/roachpb",
			"sqlfmt/cockroach/pkg/sql/catalog",
			"sqlfmt/cockroach/pkg/sql/execinfrapb",
			"sqlfmt/cockroach/pkg/sql/tree",
		}, nil,
	)
}
