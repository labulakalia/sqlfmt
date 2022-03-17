// Copyright 2020 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package sql_test

import (
	"context"
	gosql "database/sql"
	"net/url"
	"testing"

	"sqlfmt/cockroach/pkg/base"
	"sqlfmt/cockroach/pkg/ccl/kvccl/kvtenantccl"
	"sqlfmt/cockroach/pkg/sql/catalog/catconstants"
	"sqlfmt/cockroach/pkg/sql/sqltestutils"
	"sqlfmt/cockroach/pkg/testutils/serverutils"
	"sqlfmt/cockroach/pkg/testutils/skip"
	"sqlfmt/cockroach/pkg/testutils/sqlutils"
	"sqlfmt/cockroach/pkg/util/leaktest"
	"sqlfmt/cockroach/pkg/util/log"
	"github.com/stretchr/testify/require"
)

// Dummy import to pull in kvtenantccl. This allows us to start tenants.
// TODO(yuzefovich): break up the dependency on CCL code.
var _ = kvtenantccl.Connector{}

func TestTelemetry(t *testing.T) {
	defer leaktest.AfterTest(t)()
	defer log.Scope(t).Close(t)
	skip.UnderRace(t, "takes >1min under race")

	sqltestutils.TelemetryTest(
		t,
		[]base.TestServerArgs{{}},
		true, /* testTenant */
	)
}

func TestTelemetryRecordCockroachShell(t *testing.T) {
	defer leaktest.AfterTest(t)()
	defer log.Scope(t).Close(t)
	cluster := serverutils.StartNewTestCluster(
		t,
		1,
		base.TestClusterArgs{},
	)
	defer cluster.Stopper().Stop(context.Background())

	pgUrl, cleanupFn := sqlutils.PGUrl(
		t,
		cluster.Server(0).ServingSQLAddr(),
		"TestTelemetryRecordCockroachShell",
		url.User("root"),
	)
	defer cleanupFn()
	q := pgUrl.Query()

	q.Add("application_name", catconstants.ReportableAppNamePrefix+catconstants.InternalSQLAppName)
	pgUrl.RawQuery = q.Encode()

	db, err := gosql.Open("postgres", pgUrl.String())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var appName string
	err = db.QueryRow("SHOW application_name").Scan(&appName)
	require.NoError(t, err)
	require.Equal(t, catconstants.ReportableAppNamePrefix+catconstants.InternalSQLAppName, appName)

	var counter int
	err = db.QueryRow(
		"SELECT usage_count FROM crdb_internal.feature_usage WHERE feature_name = 'sql.connection.cockroach_cli'",
	).Scan(&counter)
	require.NoError(t, err)
	require.Equal(t, 1, counter)

}
