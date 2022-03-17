// Copyright 2020 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package colflow

import (
	"context"
	"math/rand"

	"sqlfmt/cockroach/pkg/col/coldata"
	"sqlfmt/cockroach/pkg/sql/colexecerror"
	"sqlfmt/cockroach/pkg/sql/colexecop"
	"sqlfmt/cockroach/pkg/util/log"
	"sqlfmt/cockroach/pkg/util/randutil"
	"github.com/cockroachdb/errors"
)

// panicInjector is a helper Operator that will randomly inject panics into
// Init and Next methods of the wrapped operator.
type panicInjector struct {
	colexecop.OneInputNode
	colexecop.InitHelper
	rng *rand.Rand
}

var _ colexecop.Operator = &panicInjector{}

const (
	// These constants were chosen arbitrarily with the guiding thought that
	// Init() methods are called less frequently, so the probability of
	// injecting in Init() should be higher. At the same time, we don't want
	// for the vectorized flows to always run into these injected panics, so
	// both numbers are relatively low.
	initPanicInjectionProbability = 0.001
	nextPanicInjectionProbability = 0.00001
)

// newPanicInjector creates a new panicInjector.
func newPanicInjector(input colexecop.Operator) colexecop.Operator {
	rng, _ := randutil.NewTestRand()
	return &panicInjector{
		OneInputNode: colexecop.OneInputNode{Input: input},
		rng:          rng,
	}
}

func (i *panicInjector) Init(ctx context.Context) {
	if !i.InitHelper.Init(ctx) {
		return
	}
	if i.rng.Float64() < initPanicInjectionProbability {
		log.Info(i.Ctx, "injecting panic in Init")
		colexecerror.ExpectedError(errors.New("injected panic in Init"))
	}
	i.Input.Init(i.Ctx)
}

func (i *panicInjector) Next() coldata.Batch {
	if i.rng.Float64() < nextPanicInjectionProbability {
		log.Info(i.Ctx, "injecting panic in Next")
		colexecerror.ExpectedError(errors.New("injected panic in Next"))
	}
	return i.Input.Next()
}
