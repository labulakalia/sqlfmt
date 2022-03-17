// Copyright 2021 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package diagnostics_test

import (
	"context"
	"net/url"
	"testing"

	"sqlfmt/cockroach/pkg/base"
	"sqlfmt/cockroach/pkg/build"
	"sqlfmt/cockroach/pkg/server"
	"sqlfmt/cockroach/pkg/server/diagnostics"
	"sqlfmt/cockroach/pkg/testutils/diagutils"
	"sqlfmt/cockroach/pkg/testutils/serverutils"
	"sqlfmt/cockroach/pkg/util/cloudinfo"
	"sqlfmt/cockroach/pkg/util/leaktest"
	"sqlfmt/cockroach/pkg/util/log"
	"github.com/stretchr/testify/require"
)

func TestCheckVersion(t *testing.T) {
	defer leaktest.AfterTest(t)()
	defer log.Scope(t).Close(t)
	defer cloudinfo.Disable()()

	ctx := context.Background()

	t.Run("expected-reporting", func(t *testing.T) {
		r := diagutils.NewServer()
		defer r.Close()

		url := r.URL()
		s, _, _ := serverutils.StartServer(t, base.TestServerArgs{
			Knobs: base.TestingKnobs{
				Server: &server.TestingKnobs{
					DiagnosticsTestingKnobs: diagnostics.TestingKnobs{
						OverrideUpdatesURL: &url,
					},
				},
			},
		})
		defer s.Stopper().Stop(ctx)
		s.UpdateChecker().(*diagnostics.UpdateChecker).CheckForUpdates(ctx)
		r.Close()

		require.Equal(t, 1, r.NumRequests())

		last := r.LastRequestData()
		require.Equal(t, s.(*server.TestServer).ClusterID().String(), last.UUID)
		require.Equal(t, "system", last.TenantID)
		require.Equal(t, build.GetInfo().Tag, last.Version)
		require.Equal(t, "OSS", last.LicenseType)
		require.Equal(t, "false", last.Internal)
	})

	t.Run("npe", func(t *testing.T) {
		// Ensure nil, which happens when an empty env override URL is used, does not
		// cause a crash.
		var nilURL *url.URL
		s, _, _ := serverutils.StartServer(t, base.TestServerArgs{
			Knobs: base.TestingKnobs{
				Server: &server.TestingKnobs{
					DiagnosticsTestingKnobs: diagnostics.TestingKnobs{
						OverrideUpdatesURL:   &nilURL,
						OverrideReportingURL: &nilURL,
					},
				},
			},
		})
		defer s.Stopper().Stop(ctx)
		s.UpdateChecker().(*diagnostics.UpdateChecker).CheckForUpdates(ctx)
		s.DiagnosticsReporter().(*diagnostics.Reporter).ReportDiagnostics(ctx)
	})
}
