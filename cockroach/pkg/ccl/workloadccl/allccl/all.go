// Copyright 2018 The Cockroach Authors.
//
// Licensed as a CockroachDB Enterprise file under the Cockroach Community
// License (the "License"); you may not use this file except in compliance with
// the License. You may obtain a copy of the License at
//
//     https://sqlfmt/cockroach/blob/master/licenses/CCL.txt

package allccl

// We import each of the workloads below, so a single import of this package
// enables registration of all workloads.

import (
	// workloads
	_ "sqlfmt/cockroach/pkg/ccl/workloadccl/roachmartccl"
	_ "sqlfmt/cockroach/pkg/workload/bank"
	_ "sqlfmt/cockroach/pkg/workload/bulkingest"
	_ "sqlfmt/cockroach/pkg/workload/connectionlatency"
	_ "sqlfmt/cockroach/pkg/workload/debug"
	_ "sqlfmt/cockroach/pkg/workload/examples"
	_ "sqlfmt/cockroach/pkg/workload/geospatial"
	_ "sqlfmt/cockroach/pkg/workload/indexes"
	_ "sqlfmt/cockroach/pkg/workload/jsonload"
	_ "sqlfmt/cockroach/pkg/workload/kv"
	_ "sqlfmt/cockroach/pkg/workload/ledger"
	_ "sqlfmt/cockroach/pkg/workload/movr"
	_ "sqlfmt/cockroach/pkg/workload/querybench"
	_ "sqlfmt/cockroach/pkg/workload/querylog"
	_ "sqlfmt/cockroach/pkg/workload/queue"
	_ "sqlfmt/cockroach/pkg/workload/rand"
	_ "sqlfmt/cockroach/pkg/workload/schemachange"
	_ "sqlfmt/cockroach/pkg/workload/sqlsmith"
	_ "sqlfmt/cockroach/pkg/workload/tpcc"
	_ "sqlfmt/cockroach/pkg/workload/tpccchecks"
	_ "sqlfmt/cockroach/pkg/workload/tpcds"
	_ "sqlfmt/cockroach/pkg/workload/tpch"
	_ "sqlfmt/cockroach/pkg/workload/ycsb"
)
