// Copyright 2017 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package logcrash_test

import (
	"context"
	"os"
	"testing"

	"github.com/labulakalia/sqlfmt/cockroach/pkg/security"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/security/securitytest"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/server"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/settings/cluster"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/testutils/serverutils"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/util/log/logcrash"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/util/randutil"
)

func TestMain(m *testing.M) {
	randutil.SeedForTests()
	security.SetAssetLoader(securitytest.EmbeddedAssets)
	serverutils.InitTestServerFactory(server.TestServerFactory)
	ctx := context.Background()

	// MakeTestingClusterSettings initializes log.ReportingSettings to this
	// instance of setting values.
	// TODO(knz): This comment appears to be untrue.
	st := cluster.MakeTestingClusterSettings()
	logcrash.DiagnosticsReportingEnabled.Override(ctx, &st.SV, false)
	logcrash.CrashReports.Override(ctx, &st.SV, false)

	os.Exit(m.Run())
}
