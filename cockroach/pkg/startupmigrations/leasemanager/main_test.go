// Copyright 2015 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package leasemanager_test

import (
	"os"
	"testing"

	"github.com/labulakalia/sqlfmt/cockroach/pkg/security"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/security/securitytest"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/server"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/testutils/serverutils"
)

func init() {
	security.SetAssetLoader(securitytest.EmbeddedAssets)
}

func TestMain(m *testing.M) {
	serverutils.InitTestServerFactory(server.TestServerFactory)
	os.Exit(m.Run())
}

//go:generate ../../util/leaktest/add-leaktest.sh *_test.go
