// Copyright 2019 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

/*
Package impl is a stub package that imports all of the concrete
implementations of the various cloud storage providers to trigger their
initialization-time registration with the cloud storage provider registry.
*/
package impl

import (
	// Import all the cloud provider packages to register them.
	_ "sqlfmt/cockroach/pkg/cloud/amazon"
	_ "sqlfmt/cockroach/pkg/cloud/azure"
	_ "sqlfmt/cockroach/pkg/cloud/gcp"
	_ "sqlfmt/cockroach/pkg/cloud/httpsink"
	_ "sqlfmt/cockroach/pkg/cloud/nodelocal"
	_ "sqlfmt/cockroach/pkg/cloud/nullsink"
	_ "sqlfmt/cockroach/pkg/cloud/userfile"
)
