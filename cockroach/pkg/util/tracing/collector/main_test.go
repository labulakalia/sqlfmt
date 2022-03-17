// Copyright 2021 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package collector_test

import (
	"os"
	"testing"

	"sqlfmt/cockroach/pkg/ccl/utilccl"
	"sqlfmt/cockroach/pkg/security"
	"sqlfmt/cockroach/pkg/security/securitytest"
	"sqlfmt/cockroach/pkg/server"
	"sqlfmt/cockroach/pkg/testutils/serverutils"
	"sqlfmt/cockroach/pkg/util/randutil"
)

func TestMain(m *testing.M) {
	defer utilccl.TestingEnableEnterprise()()
	security.SetAssetLoader(securitytest.EmbeddedAssets)
	randutil.SeedForTests()
	serverutils.InitTestServerFactory(server.TestServerFactory)
	os.Exit(m.Run())
}

//go:generate ../../leaktest/add-leaktest.sh *_test.go
