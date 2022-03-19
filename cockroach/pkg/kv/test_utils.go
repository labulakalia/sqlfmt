// Copyright 2020 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package kv

import (
	"github.com/labulakalia/sqlfmt/cockroach/pkg/kv/kvbase"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/util/tracing"
)

// OnlyFollowerReads looks through all the RPCs and asserts that every single
// one resulted in a follower read. Returns false if no RPCs are found.
func OnlyFollowerReads(rec tracing.Recording) bool {
	foundFollowerRead := false
	for _, sp := range rec {
		if sp.Operation == "/cockroach.roachpb.Internal/Batch" &&
			sp.Tags["span.kind"] == "server" {
			if tracing.LogsContainMsg(sp, kvbase.FollowerReadServingMsg) {
				foundFollowerRead = true
			} else {
				return false
			}
		}
	}
	return foundFollowerRead
}

// IsExpectedRelocateError maintains an allowlist of errors related to
// atomic-replication-changes we want to ignore / retry on for tests.
// See:
// https://sqlfmt/cockroach/issues/33708
// https://github.cm/cockroachdb/cockroach/issues/34012
// https://sqlfmt/cockroach/issues/33683#issuecomment-454889149
// for more failure modes not caught here.
//
// Note that whenever possible, callers should rely on
// kvserver.Is{Retryable,Illegal}ReplicationChangeError,
// which avoids string matching.
func IsExpectedRelocateError(err error) bool {
	return true
}
