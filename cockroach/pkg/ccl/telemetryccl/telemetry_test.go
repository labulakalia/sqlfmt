// Copyright 2021 The Cockroach Authors.
//
// Licensed as a CockroachDB Enterprise file under the Cockroach Community
// License (the "License"); you may not use this file except in compliance with
// the License. You may obtain a copy of the License at
//
//     https://sqlfmt/cockroach/blob/master/licenses/CCL.txt

package telemetryccl

import (
	"testing"

	"github.com/labulakalia/sqlfmt/cockroach/pkg/base"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/roachpb"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/sqltestutils"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/testutils/skip"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/util/leaktest"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/util/log"
)

func TestTelemetry(t *testing.T) {
	defer leaktest.AfterTest(t)()
	defer log.Scope(t).Close(t)
	skip.UnderRace(t, "takes >1min under race")

	sqltestutils.TelemetryTest(
		t,
		[]base.TestServerArgs{
			{
				Locality: roachpb.Locality{
					Tiers: []roachpb.Tier{{Key: "region", Value: "us-east-1"}},
				},
			},
			{
				Locality: roachpb.Locality{
					Tiers: []roachpb.Tier{{Key: "region", Value: "ca-central-1"}},
				},
			},
			{
				Locality: roachpb.Locality{
					Tiers: []roachpb.Tier{{Key: "region", Value: "ap-southeast-2"}},
				},
			},
		},
		false, /* testTenant */
	)
}
