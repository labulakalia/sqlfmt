// Copyright 2020 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// Package slprovider exposes an implementation of the sqlliveness.Provider
// interface.
package slprovider

import (
	"context"

	"sqlfmt/cockroach/pkg/keys"
	"sqlfmt/cockroach/pkg/kv"
	"sqlfmt/cockroach/pkg/settings/cluster"
	"sqlfmt/cockroach/pkg/sql/sqlliveness"
	"sqlfmt/cockroach/pkg/sql/sqlliveness/slinstance"
	"sqlfmt/cockroach/pkg/sql/sqlliveness/slstorage"
	"sqlfmt/cockroach/pkg/util/hlc"
	"sqlfmt/cockroach/pkg/util/log"
	"sqlfmt/cockroach/pkg/util/metric"
	"sqlfmt/cockroach/pkg/util/stop"
)

// New constructs a new Provider.
func New(
	ambientCtx log.AmbientContext,
	stopper *stop.Stopper,
	clock *hlc.Clock,
	db *kv.DB,
	codec keys.SQLCodec,
	settings *cluster.Settings,
	testingKnobs *sqlliveness.TestingKnobs,
) sqlliveness.Provider {
	storage := slstorage.NewStorage(ambientCtx, stopper, clock, db, codec, settings)
	instance := slinstance.NewSQLInstance(stopper, clock, storage, settings, testingKnobs)
	return &provider{
		Storage:  storage,
		Instance: instance,
	}
}

func (p *provider) Start(ctx context.Context) {
	p.Storage.Start(ctx)
	p.Instance.Start(ctx)
}

func (p *provider) Metrics() metric.Struct {
	return p.Storage.Metrics()
}

type provider struct {
	*slstorage.Storage
	*slinstance.Instance
}
