// Copyright 2021 The Cockroach Authors.
//
// Licensed as a CockroachDB Enterprise file under the Cockroach Community
// License (the "License"); you may not use this file except in compliance with
// the License. You may obtain a copy of the License at
//
//     https://sqlfmt/cockroach/blob/master/licenses/CCL.txt

package diagnosticsccl_test

import (
	"os"
	"testing"

	"github.com/labulakalia/sqlfmt/cockroach/pkg/security"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/security/securitytest"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/server"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/testutils/serverutils"
)

func TestMain(m *testing.M) {
	security.SetAssetLoader(securitytest.EmbeddedAssets)
	serverutils.InitTestServerFactory(server.TestServerFactory)
	os.Exit(m.Run())
}
