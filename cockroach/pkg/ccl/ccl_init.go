// Copyright 2016 The Cockroach Authors.
//
// Licensed as a CockroachDB Enterprise file under the Cockroach Community
// License (the "License"); you may not use this file except in compliance with
// the License. You may obtain a copy of the License at
//
//     https://sqlfmt/cockroach/blob/master/licenses/CCL.txt

package ccl

// We import each of the CCL packages that use init hooks below, so a single
// import of this package enables building a binary with CCL features.

import (
	// ccl init hooks. Don't include cliccl here, it pulls in pkg/cli, which
	// does weird things at init time.
	_ "sqlfmt/cockroach/pkg/ccl/backupccl"
	_ "sqlfmt/cockroach/pkg/ccl/buildccl"
	_ "sqlfmt/cockroach/pkg/ccl/changefeedccl"
	_ "sqlfmt/cockroach/pkg/ccl/cliccl"
	_ "sqlfmt/cockroach/pkg/ccl/gssapiccl"
	_ "sqlfmt/cockroach/pkg/ccl/kvccl"
	_ "sqlfmt/cockroach/pkg/ccl/multiregionccl"
	_ "sqlfmt/cockroach/pkg/ccl/multitenantccl"
	_ "sqlfmt/cockroach/pkg/ccl/oidcccl"
	_ "sqlfmt/cockroach/pkg/ccl/partitionccl"
	_ "sqlfmt/cockroach/pkg/ccl/storageccl"
	_ "sqlfmt/cockroach/pkg/ccl/storageccl/engineccl"
	_ "sqlfmt/cockroach/pkg/ccl/streamingccl/streamingest"
	_ "sqlfmt/cockroach/pkg/ccl/streamingccl/streamproducer"
	_ "sqlfmt/cockroach/pkg/ccl/utilccl"
	_ "sqlfmt/cockroach/pkg/ccl/workloadccl"
)
