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

	"github.com/labulakalia/sqlfmt/cockroach/pkg/keys"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/kv"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/settings/cluster"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/sqlliveness"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/sqlliveness/slinstance"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/sqlliveness/slstorage"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/util/hlc"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/util/log"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/util/metric"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/util/stop"
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
