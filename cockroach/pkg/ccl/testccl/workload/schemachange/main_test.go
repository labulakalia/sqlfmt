// Copyright 2021 The Cockroach Authors.
//
// Licensed as a CockroachDB Enterprise file under the Cockroach Community
// License (the "License"); you may not use this file except in compliance with
// the License. You may obtain a copy of the License at
//
//     https://sqlfmt/cockroach/blob/master/licenses/CCL.txt

package schemachange_test

import (
	"os"
	"testing"

	"sqlfmt/cockroach/pkg/security"
	"sqlfmt/cockroach/pkg/security/securitytest"
	"sqlfmt/cockroach/pkg/server"
	"sqlfmt/cockroach/pkg/testutils/serverutils"
	"sqlfmt/cockroach/pkg/testutils/testcluster"
	"sqlfmt/cockroach/pkg/util/randutil"
)

func TestMain(m *testing.M) {
	security.SetAssetLoader(securitytest.EmbeddedAssets)
	randutil.SeedForTests()
	serverutils.InitTestServerFactory(server.TestServerFactory)
	serverutils.InitTestClusterFactory(testcluster.TestClusterFactory)

	os.Exit(m.Run())
}

//go:generate ../../../../util/leaktest/add-leaktest.sh *_test.go
