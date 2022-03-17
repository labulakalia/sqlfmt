// Copyright 2015 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package kvcoord_test

import (
	"os"
	"testing"

	"sqlfmt/cockroach/pkg/security"
	"sqlfmt/cockroach/pkg/security/securitytest"
	"sqlfmt/cockroach/pkg/server"
	"sqlfmt/cockroach/pkg/testutils/buildutil"
	"sqlfmt/cockroach/pkg/testutils/serverutils"
	"sqlfmt/cockroach/pkg/util/leaktest"
	"sqlfmt/cockroach/pkg/util/log"
	"sqlfmt/cockroach/pkg/util/randutil"
)

//go:generate ../../../util/leaktest/add-leaktest.sh *_test.go

func init() {
	security.SetAssetLoader(securitytest.EmbeddedAssets)
}

func TestForbiddenDeps(t *testing.T) {
	defer leaktest.AfterTest(t)()
	defer log.Scope(t).Close(t)
	// Verify kv does not depend on storage (or any of its subpackages).
	buildutil.VerifyNoImports(t,
		"sqlfmt/cockroach/pkg/kv", true,
		// TODO(tschottdorf): should really disallow ./storage/... but at the
		// time of writing there's a (legit) dependency on `enginepb`.
		[]string{
			"sqlfmt/cockroach/pkg/storage",
			"sqlfmt/cockroach/pkg/storage/engine",
		},
		[]string{})
}

func TestMain(m *testing.M) {
	serverutils.InitTestServerFactory(server.TestServerFactory)
	randutil.SeedForTests()
	os.Exit(m.Run())
}
